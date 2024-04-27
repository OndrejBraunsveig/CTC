package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

var numCars int
var gas_cars, diesel_cars, lpg_cars, electric_cars int
var gas_total, diesel_total, lpg_total, electric_total, register_total time.Duration
var gas_max, diesel_max, lpg_max, electric_max, register_max time.Duration

type Config struct {
	Cars struct {
		Count     int           `yaml:"count"`
		ArivalMin time.Duration `yaml:"arrival_time_min"`
		ArivalMax time.Duration `yaml:"arrival_time_max"`
	} `yaml:"cars"`
	Stations struct {
		Gas struct {
			Count    int           `yaml:"count"`
			ServeMin time.Duration `yaml:"serve_time_min"`
			ServeMax time.Duration `yaml:"serve_time_max"`
		} `yaml:"gas"`
		Diesel struct {
			Count    int           `yaml:"count"`
			ServeMin time.Duration `yaml:"serve_time_min"`
			ServeMax time.Duration `yaml:"serve_time_max"`
		} `yaml:"diesel"`
		Lpg struct {
			Count    int           `yaml:"count"`
			ServeMin time.Duration `yaml:"serve_time_min"`
			ServeMax time.Duration `yaml:"serve_time_max"`
		} `yaml:"lpg"`
		Electric struct {
			Count    int           `yaml:"count"`
			ServeMin time.Duration `yaml:"serve_time_min"`
			ServeMax time.Duration `yaml:"serve_time_max"`
		} `yaml:"electric"`
	} `yaml:"stations"`
	Registers struct {
		Count     int           `yaml:"count"`
		HandleMin time.Duration `yaml:"handle_time_min"`
		HandleMax time.Duration `yaml:"handle_time_max"`
	} `yaml:"registers"`
}

type Car struct {
	QueueStart      time.Time
	ServeQueueTime  time.Duration
	HandleQueueTime time.Duration
	StationTime     time.Duration
	RegisterTime    time.Duration
}

type Station struct {
	Type      string
	Count     int
	ServeTime [2]time.Duration
}

type CashRegister struct {
	Count      int
	HandleTime [2]time.Duration
}

func readConfig() *Config {
	f, err := os.Open("config.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	config := Config{}
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		panic(err)
	}

	return &config
}

// Serves car at the station
func serveCar(car Car, station Station, registerQueue chan Car) {

	serveTime := time.Duration(rand.Intn(int(station.ServeTime[1])+1-int(station.ServeTime[0])) + int(station.ServeTime[0]))
	time.Sleep(serveTime)

	// Station stats update
	queueTime := time.Since(car.QueueStart) - serveTime
	updateStationStats(queueTime, station)

	car.ServeQueueTime = queueTime
	car.StationTime = serveTime
	car.QueueStart = time.Now()

	registerQueue <- car
}

// Handles car at the register
func handleCar(car Car, cashRegister CashRegister, wg *sync.WaitGroup) {
	defer wg.Done()

	handleTime := time.Duration(rand.Intn(int(cashRegister.HandleTime[1])+1-int(cashRegister.HandleTime[0])) + int(cashRegister.HandleTime[0]))
	time.Sleep(handleTime)

	// Register stats update
	queueTime := time.Since(car.QueueStart) - handleTime
	register_total += queueTime
	if queueTime > register_max {
		register_max = queueTime
	}

	car.HandleQueueTime = queueTime
	car.RegisterTime = handleTime
}

func updateStationStats(queueTime time.Duration, station Station) {
	switch {
	case station.Type == "gas":
		gas_cars += 1
		gas_total += queueTime
		if queueTime > gas_max {
			gas_max = queueTime
		}
	case station.Type == "diesel":
		diesel_cars += 1
		diesel_total += queueTime
		if queueTime > diesel_max {
			diesel_max = queueTime
		}
	case station.Type == "lpg":
		lpg_cars += 1
		lpg_total += queueTime
		if queueTime > lpg_max {
			lpg_max = queueTime
		}
	case station.Type == "electric":
		electric_cars += 1
		electric_total += queueTime
		if queueTime > electric_max {
			electric_max = queueTime
		}
	}
}

func printStats() {
	fmt.Printf("Stats of the simulation:\n")
	fmt.Printf("\nstations:\n")
	fmt.Printf("  gas:\n")
	fmt.Printf("    total_cars: %d\n", gas_cars)
	fmt.Printf("    total_time: %v\n", gas_total)
	fmt.Printf("    avg_queue_time: %v\n", gas_total/time.Duration(gas_cars))
	fmt.Printf("    max_queue_time: %v\n", gas_max)
	fmt.Printf("  diesel:\n")
	fmt.Printf("    total_cars: %d\n", diesel_cars)
	fmt.Printf("    total_time: %v\n", diesel_total)
	fmt.Printf("    avg_queue_time: %v\n", diesel_total/time.Duration(diesel_cars))
	fmt.Printf("    max_queue_time: %v\n", diesel_max)
	fmt.Printf("  lpg:\n")
	fmt.Printf("    total_cars: %d\n", lpg_cars)
	fmt.Printf("    total_time: %v\n", lpg_total)
	fmt.Printf("    avg_queue_time: %v\n", lpg_total/time.Duration(lpg_cars))
	fmt.Printf("    max_queue_time: %v\n", lpg_max)
	fmt.Printf("  electric:\n")
	fmt.Printf("    total_cars: %d\n", electric_cars)
	fmt.Printf("    total_time: %v\n", electric_total)
	fmt.Printf("    avg_queue_time: %v\n", electric_total/time.Duration(electric_cars))
	fmt.Printf("    max_queue_time: %v\n", electric_max)
	fmt.Printf("registers:\n")
	fmt.Printf("  total_cars: %d\n", numCars)
	fmt.Printf("  total_time: %v\n", register_total)
	fmt.Printf("  avg_queue_time: %v\n", register_total/time.Duration(numCars))
	fmt.Printf("  max_queue_time: %v\n", register_max)
}

func main() {

	// Initial configuration
	config := readConfig()

	numCars = config.Cars.Count

	stations := []Station{
		{"gas", config.Stations.Gas.Count, [2]time.Duration{config.Stations.Gas.ServeMin, config.Stations.Gas.ServeMax}},
		{"diesel", config.Stations.Diesel.Count, [2]time.Duration{config.Stations.Diesel.ServeMin, config.Stations.Diesel.ServeMax}},
		{"lpg", config.Stations.Lpg.Count, [2]time.Duration{config.Stations.Lpg.ServeMin, config.Stations.Lpg.ServeMax}},
		{"electric", config.Stations.Electric.Count, [2]time.Duration{config.Stations.Electric.ServeMin, config.Stations.Electric.ServeMax}},
	}

	register := CashRegister{config.Registers.Count, [2]time.Duration{config.Registers.HandleMin, config.Registers.HandleMax}}

	var wg sync.WaitGroup
	wg.Add(numCars)

	// Creation of station and register channels
	stationQueue := make(chan Car, numCars)
	registerQueue := make(chan Car, numCars)

	// Station goroutines start
	for _, station := range stations {
		for i := 0; i < station.Count; i++ {
			go func(station Station) {
				for {
					if len(stationQueue) != 0 {
						car := <-stationQueue
						serveCar(car, station, registerQueue)
					}
				}
			}(station)
		}
	}

	// Register goroutines start
	for i := 0; i < register.Count; i++ {
		go func(register CashRegister) {
			for {
				if len(registerQueue) != 0 {
					car := <-registerQueue
					handleCar(car, register, &wg)
				}
			}
		}(register)
	}

	fmt.Printf("\nSimulation with %v cars in progress . . .\n", numCars)

	// After stations and registers start, cars begin their arrival
	for i := 0; i < numCars; i++ {
		car := Car{
			QueueStart: time.Now(),
		}
		stationQueue <- car
		time.Sleep(time.Duration(rand.Intn(int(config.Cars.ArivalMax)+1-int(config.Cars.ArivalMin)) + int(config.Cars.ArivalMin)))
	}

	wg.Wait()
	close(stationQueue)
	close(registerQueue)

	fmt.Printf("\nSimulation done!\n")
	printStats()
}
