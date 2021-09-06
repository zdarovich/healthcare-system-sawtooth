package main

import (
	"context"
	"fmt"
	"github.com/robfig/cron"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"healthcare-system-sawtooth/client/db/models"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	cronRunner := cron.New()

	cronRunner.AddFunc("@every 10s", func() {
		fmt.Println("cleanup")
		ctx := context.Background()
		now := time.Now().Unix()
		datas, err := models.GetDataByExpiration(ctx, now)
		if err != nil {
			log.Println(err)
			return
		}
		if len(datas) == 0 {
			return
		}
		var oids []*primitive.ObjectID
		for _, d := range datas {
			oids = append(oids, d.OID)
		}
		err = models.DeleteDatasByOid(oids)
		if err != nil {
			log.Println(err)
			return
		}
	})
	cronRunner.Start()
	// Shutdown.
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
}
