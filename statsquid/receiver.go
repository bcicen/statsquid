package main

import (
	"fmt"
	"math"

	"gopkg.in/redis.v3"
)

func Round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}

func readIn() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}
	pubsub, err := client.Subscribe("statsquid")
	if err != nil {
		panic(err)
	}
	defer pubsub.Close()

	var stat StatSquidStat
	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			panic(err)
		}
		stat.Unpack(msg.Payload)
		fmt.Println(stat.Names)
	}
}
