package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firestorepb "cloud.google.com/go/firestore/apiv1/firestorepb"
	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/api/option"
)

var ctx = context.Background()
var rdb *redis.Client
var client *firestore.Client

func Hello(w http.ResponseWriter, r *http.Request) {
	// Cookie からセッショントークンを取得
	tok, err := r.Cookie("next-auth.session-token")

	// Cookieにセッショントークンがあるかチェック
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized!!!")
		return
	}

	// Redisにトークンがあるかチェック
	_, err = rdb.Get(ctx, "user:session:"+tok.Value).Result()
	if err == redis.Nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized!!!")
		return
	}

	log.Println(tok.Value)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello!!!")
}

func Firestore(w http.ResponseWriter, r *http.Request) {
	// Cookie からセッショントークンを取得
	tok, err := r.Cookie("next-auth.session-token")

	// Cookieにセッショントークンがあるかチェック
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized!!!")
		return
	}

	// Firestoreにトークンがあるかチェック
	query := client.Collection("sessions").Where("sessionToken", "==", tok.Value).Limit(1)
	aggregationQuery := query.NewAggregationQuery().WithCount("all")
	results, err := aggregationQuery.Get(ctx)
	if err != nil {
		return
	}

	count, ok := results["all"]
	if !ok {
		return
	}

	countValue := count.(*firestorepb.Value)
	if countValue.GetIntegerValue() == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "Unauthorized!!!")
		return
	}

	log.Println(tok.Value)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello!!!")
}

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Panicln("not found .env.local")
	}
	opt, _ := redis.ParseURL(os.Getenv("UPSTASH_REDIS_PARSE_URL"))
	rdb = redis.NewClient(opt)

	sa := option.WithCredentialsFile("token.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s := time.Now()
			Hello(w, r)
			fmt.Printf("process time: %s\n", time.Since(s))
		}
	})

	http.HandleFunc("/firestore", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s := time.Now()
			Firestore(w, r)
			fmt.Printf("process time: %s\n", time.Since(s))
		}
	})
	log.Fatal(http.ListenAndServe(":8000", nil))
}
