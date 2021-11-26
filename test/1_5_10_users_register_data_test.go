package test

import (
	"encoding/csv"
	"fmt"
	"github.com/jamiealquiza/tachymeter"
	"healthcare-system-sawtooth/client/lib"
	"healthcare-system-sawtooth/client/user"
	"log"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

func Test_1_User_Register(t *testing.T) {
	requestSamples := 100
	concurrentClients := 1

	reportFileName := fmt.Sprintf("report/%s-Test_1_User_Register.csv", time.Now().Format("2006-01-02T15:04:05"))
	file, err := os.Create(reportFileName)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.Write([]string{
		"cumulative_time(seconds)",
		"latency(seconds)",
		"transaction_per_second(tps)",
		"memory_load(bytes)",
		"send_rate(bytes/seconds)",
	})
	if err != nil {
		t.Error(err)
	}
	writer.Flush()

	// Start wall time for all Goroutines.
	var wg1 sync.WaitGroup

	memory := 10
	for sample := 0; sample < requestSamples; sample++ {
		wallTimeStart := time.Now()

		c := tachymeter.New(&tachymeter.Config{Size: concurrentClients * requestSamples})
		for i := 0; i < concurrentClients; i++ {
			wg1.Add(1)
			go patientRegister(t, c, &wg1, memory)
		}
		wg1.Wait()

		c.SetWallTime(time.Since(wallTimeStart))
		data := []string{
			fmt.Sprintf("%f", c.Calc().Time.Cumulative.Seconds()),
			fmt.Sprintf("%f", c.Calc().Time.Avg.Seconds()),
			fmt.Sprintf("%f", c.Calc().Rate.Second),
			fmt.Sprintf("%d", memory),
			fmt.Sprintf("%f", float64(memory*concurrentClients)/c.Calc().Time.Cumulative.Seconds()),
		}
		log.Printf("data %+v \n", data)

		err := writer.Write(data)
		if err != nil {
			t.Error(err)
		}
		log.Printf("sent as bytes %d \n", memory)
		writer.Flush()
	}
}

func Test_5_User_Register(t *testing.T) {
	requestSamples := 100
	concurrentClients := 5

	reportFileName := fmt.Sprintf("report/%s-Test_5_User_Register.csv", time.Now().Format("2006-01-02T15:04:05"))
	file, err := os.Create(reportFileName)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.Write([]string{
		"cumulative_time(seconds)",
		"latency(seconds)",
		"transaction_per_second(tps)",
		"memory_load(bytes)",
		"send_rate(bytes/seconds)",
	})
	if err != nil {
		t.Error(err)
	}
	writer.Flush()

	// Start wall time for all Goroutines.
	var wg1 sync.WaitGroup

	memory := 10
	for sample := 0; sample < requestSamples; sample++ {
		wallTimeStart := time.Now()

		c := tachymeter.New(&tachymeter.Config{Size: concurrentClients * requestSamples})
		for i := 0; i < concurrentClients; i++ {
			wg1.Add(1)
			go patientRegister(t, c, &wg1, memory)
		}
		wg1.Wait()

		c.SetWallTime(time.Since(wallTimeStart))
		data := []string{
			fmt.Sprintf("%f", c.Calc().Time.Cumulative.Seconds()),
			fmt.Sprintf("%f", c.Calc().Time.Avg.Seconds()),
			fmt.Sprintf("%f", c.Calc().Rate.Second),
			fmt.Sprintf("%d", memory),
			fmt.Sprintf("%f", float64(memory*concurrentClients)/c.Calc().Time.Cumulative.Seconds()),
		}
		log.Printf("data %+v \n", data)

		err := writer.Write(data)
		if err != nil {
			t.Error(err)
		}
		log.Printf("sent as bytes %d \n", memory)
		writer.Flush()
	}
}

func Test_10_User_Register(t *testing.T) {
	requestSamples := 100
	concurrentClients := 3

	reportFileName := fmt.Sprintf("report/%s-Test_3_User_Register.csv", time.Now().Format("2006-01-02T15:04:05"))
	file, err := os.Create(reportFileName)
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.Write([]string{
		"cumulative_time(seconds)",
		"latency(seconds)",
		"transaction_per_second(tps)",
		"memory_load(bytes)",
		"send_rate(bytes/seconds)",
	})
	if err != nil {
		t.Error(err)
	}
	writer.Flush()

	// Start wall time for all Goroutines.
	var wg1 sync.WaitGroup

	memory := 10
	for sample := 0; sample < requestSamples; sample++ {
		wallTimeStart := time.Now()

		c := tachymeter.New(&tachymeter.Config{Size: concurrentClients * requestSamples})
		for i := 0; i < concurrentClients; i++ {
			wg1.Add(1)
			go patientRegister(t, c, &wg1, memory)
		}
		wg1.Wait()

		c.SetWallTime(time.Since(wallTimeStart))
		data := []string{
			fmt.Sprintf("%f", c.Calc().Time.Cumulative.Seconds()),
			fmt.Sprintf("%f", c.Calc().Time.Avg.Seconds()),
			fmt.Sprintf("%f", c.Calc().Rate.Second),
			fmt.Sprintf("%d", memory),
			fmt.Sprintf("%f", float64(memory*concurrentClients)/c.Calc().Time.Cumulative.Seconds()),
		}
		log.Printf("data %+v \n", data)

		err := writer.Write(data)
		if err != nil {
			t.Error(err)
		}
		log.Printf("sent as bytes %d \n", memory)
		writer.Flush()
	}
}

func patientRegister(t *testing.T, stats *tachymeter.Tachymeter, wg *sync.WaitGroup, patientNameLength int) {
	defer wg.Done()
	testKeyPath := "resources/keys"
	start := time.Now()

	randName := RandStringRunes(patientNameLength)
	lib.GenerateKey(randName, testKeyPath)
	privKeyPath := path.Join(testKeyPath, randName+".priv")
	cli, err := user.NewUserClient(randName, privKeyPath)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.UserRegister()
	if err != nil {
		t.Fatal(err)
	}
	stats.AddTime(time.Since(start))
	log.Printf("%s register stopped \n", randName)
}
