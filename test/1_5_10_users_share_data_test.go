package test

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
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

type UserKeyMap struct {
	Username, Keypath string
}

func initialize(t *testing.T, n int) ([]UserKeyMap, []UserKeyMap) {
	testKeyPath := "resources/keys"

	testPatientClients := make([]UserKeyMap, 0)
	testDoctorClients := make([]UserKeyMap, 0)
	for i := 0; i < n; i++ {
		randName := uuid.New().String()
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
		testPatientClients = append(testPatientClients, UserKeyMap{randName, privKeyPath})

		testDoctorClient := make(map[string]string)
		randName = uuid.New().String()
		lib.GenerateKey(randName, testKeyPath)
		privKeyPath = path.Join(testKeyPath, randName+".priv")
		cli, err = user.NewUserClient(randName, privKeyPath)
		if err != nil {
			t.Fatal(err)
		}
		testDoctorClient[randName] = privKeyPath
		err = cli.UserRegister()
		if err != nil {
			t.Fatal(err)
		}
		testDoctorClients = append(testDoctorClients, UserKeyMap{randName, privKeyPath})

	}
	return testPatientClients, testDoctorClients
}

func Test_1_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes(t *testing.T) {
	requestSamples := 100
	concurrentClients := 1
	// 1000000 = 0.95 mb
	totalBytesSentLimit := 1000000

	patients, doctors := initialize(t, concurrentClients)

	reportFileName := fmt.Sprintf("report/%s-Test_1_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes.csv", time.Now().Format("2006-01-02T15:04:05"))
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

	increaseBytes := totalBytesSentLimit / requestSamples

	log.Printf("patients %d doctors %d  \n", len(patients), len(doctors))

	for memory := increaseBytes; memory < totalBytesSentLimit; memory += increaseBytes {
		wallTimeStart := time.Now()

		c := tachymeter.New(&tachymeter.Config{Size: concurrentClients * requestSamples})
		for i, v := range patients {
			wg1.Add(1)
			go patientCreateAndShareData(t, c, &wg1, memory, v.Username, v.Keypath, doctors[0].Username, i)
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

func Test_5_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes(t *testing.T) {
	requestSamples := 100
	concurrentClients := 5
	// 1000000 = 0.95 mb
	totalBytesSentLimit := 1000000

	patients, doctors := initialize(t, concurrentClients)

	reportFileName := fmt.Sprintf("report/%s-Test_5_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes.csv", time.Now().Format("2006-01-02T15:04:05"))
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

	// 1000000 = 0.95 mb
	increaseBytes := totalBytesSentLimit / requestSamples

	log.Printf("patients %d doctors %d  \n", len(patients), len(doctors))

	for memory := increaseBytes; memory < totalBytesSentLimit; memory += increaseBytes {
		wallTimeStart := time.Now()

		c := tachymeter.New(&tachymeter.Config{Size: concurrentClients * requestSamples})
		for i, v := range patients {
			wg1.Add(1)
			go patientCreateAndShareData(t, c, &wg1, memory, v.Username, v.Keypath, doctors[0].Username, i)
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

func Test_10_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes(t *testing.T) {
	requestSamples := 100
	concurrentClients := 3
	// 1000000 = 0.95 mb
	totalBytesSentLimit := 1000000

	patients, doctors := initialize(t, concurrentClients)

	reportFileName := fmt.Sprintf("report/%s-Test_3_User_Creates_Concurently_Shares_Data_Up_To_100000_bytes.csv", time.Now().Format("2006-01-02T15:04:05"))
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

	// 1000000 = 0.95 mb
	increaseBytes := totalBytesSentLimit / requestSamples

	log.Printf("patients %d doctors %d  \n", len(patients), len(doctors))

	for memory := increaseBytes; memory < totalBytesSentLimit; memory += increaseBytes {
		wallTimeStart := time.Now()

		c := tachymeter.New(&tachymeter.Config{Size: concurrentClients * requestSamples})
		for i, v := range patients {
			wg1.Add(1)
			go patientCreateAndShareData(t, c, &wg1, memory, v.Username, v.Keypath, doctors[0].Username, i)
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

func patientCreateAndShareData(t *testing.T, stats *tachymeter.Tachymeter, wg *sync.WaitGroup, memory int, patientName, patientKeyPath,
	doctorName string, clientNumber int) {
	defer wg.Done()

	log.Printf("%s started creating and sharing data for %s as number %d \n", patientName, doctorName, clientNumber)
	cli, err := user.NewUserClient(patientName, patientKeyPath)
	if err != nil {
		t.Fatal(err)
	}

	dataName := uuid.New().String()
	data := RandStringRunes(memory)

	start := time.Now()
	dataInfo, err := cli.CreatePatientData(dataName, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = cli.ShareData(dataInfo.Hash, doctorName)
	if err != nil {
		t.Error(err)
	}
	stats.AddTime(time.Since(start))
	log.Printf("%s stopped as number %d \n", patientName, clientNumber)
}
