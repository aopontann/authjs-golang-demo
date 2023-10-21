package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

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

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Panicln("not found .env.local")
	}
	opt, _ := redis.ParseURL(os.Getenv("UPSTASH_REDIS_PARSE_URL"))
	rdb = redis.NewClient(opt)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			s := time.Now()
			Hello(w, r)
			fmt.Printf("process time: %s\n", time.Since(s))
		}
	})
	log.Fatal(http.ListenAndServe(":8000", nil))
}
