package test

import (
	"github.com/sirupsen/logrus"
	"healthcare-system-sawtooth/client/lib"
	"io/ioutil"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

var (
	requestSamples100 = 100
	letterRunes       = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	rand.Seed(time.Now().UnixNano())
	lib.Logger = logrus.New()
	lib.Logger.Out = ioutil.Discard
	lib.MongoDbUrl = "mongodb://mongodb:27017"
	lib.ValidatorURL = lib.DefaultValidatorURL
	lib.TPURL = lib.DefaultTPURL
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func calcMean(nums []float64) float64 {
	if len(nums) == 0 {
		return 0.0
	}

	var total float64
	length := len(nums)
	for _, item := range nums {
		total += item
	}
	mean := total / float64(length)
	return mean
}

func calcMedian(nums []float64) float64 {
	if len(nums) == 0 {
		return 0.0
	}

	sort.Float64s(nums)

	isEven := len(nums)%2 == 0
	mNumber := len(nums) / 2

	if !isEven {
		return nums[mNumber]
	}
	return (float64(nums[mNumber-1]) + float64(nums[mNumber])) / 2
}

func calcVarience(nums []float64) float64 {
	if len(nums) == 0 {
		return 0.0
	}

	var variance float64
	mean := calcMean(nums)

	for index := range nums {
		variance += math.Pow(nums[index]-mean, 2)
	}

	return variance / float64(len(nums))
}

func calcStdDev(nums []float64) float64 {
	if len(nums) == 0 {
		return 0.0
	}

	variance := calcVarience(nums)
	return math.Sqrt(variance)
}

func calc95Percentile(nums []float64) string {
	sort.Float64s(nums)
	nineFive := float64(len(nums)-1) * 0.95

	newSlice := nums[int(nineFive):]
	//return strconv.Itoa(newSlice[0])
	return strconv.FormatFloat(newSlice[0], 'f', 0, 64)
}

func requestsPerSecond(request int, duration time.Duration) float64 {
	convertedDuration := float64(duration) / float64(time.Second)
	toS := strconv.FormatFloat(float64(request)/convertedDuration, 'f', 2, 64)
	return stringToFloat(toS)
}
func requestsPerSecondFloat64(request int, duration float64) float64 {
	convertedDuration := duration / float64(time.Second)
	toS := strconv.FormatFloat(float64(request)/convertedDuration, 'f', 2, 64)
	return stringToFloat(toS)
}
func failedRequests(slice []int) int {
	non200 := 0

	for _, item := range slice {
		if item > 226 {
			non200++
		}
	}
	return non200
}

func stringToFloat(input string) float64 {
	if s, err := strconv.ParseFloat(input, 64); err == nil {
		return s
	}
	return 0.00
}
