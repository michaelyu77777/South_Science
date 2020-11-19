package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rifflock/lfshook"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"./objects"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs" //Log寫入設定
	"github.com/sirupsen/logrus"
)

var log_info *logrus.Logger
var log_err *logrus.Logger

//設定檔
var config objects.Config = objects.Config{}

//初始化 Log
func init() {

	//初始化Log
	init_Log()

	//初始化config
	init_config()
}

func init_Log() {

	fmt.Println("執行init()初始化")

	/**設定LOG檔層級與輸出格式*/
	//使用Info層級
	path := "./log/info/info"
	writer, _ := rotatelogs.New(
		path+".%Y%m%d%H",                            // 檔名格式
		rotatelogs.WithLinkName(path),               // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(10080*time.Minute),    // 文件最大保存時間(保留七天)
		rotatelogs.WithRotationTime(60*time.Minute), // 日誌切割時間間隔(一小時存一個檔案)
	)

	// 設定LOG等級
	pathMap := lfshook.WriterMap{
		logrus.InfoLevel: writer,
		//logrus.PanicLevel: writer, //若執行發生錯誤則會停止不進行下去
	}

	log_info = logrus.New()
	log_info.Hooks.Add(lfshook.NewHook(pathMap, &logrus.JSONFormatter{})) //Log檔綁訂相關設定

	fmt.Println("結束Info等級設定")
	log_info.Info("結束Info等級設定")

	//Error層級
	path = "./log/err/err"
	writer, _ = rotatelogs.New(
		path+".%Y%m%d%H",                            // 檔名格式
		rotatelogs.WithLinkName(path),               // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(10080*time.Minute),    // 文件最大保存時間(保留七天)
		rotatelogs.WithRotationTime(60*time.Minute), // 日誌切割時間間隔(一小時存一個檔案)
	)

	// 設定LOG等級
	pathMap = lfshook.WriterMap{
		//logrus.InfoLevel: writer,
		logrus.ErrorLevel: writer,
		//logrus.PanicLevel: writer, //若執行發生錯誤則會停止不進行下去
	}

	log_err = logrus.New()
	log_err.Hooks.Add(lfshook.NewHook(pathMap, &logrus.JSONFormatter{})) //Log檔綁訂相關設定

	fmt.Println("結束Error等級設定")
	log_info.Info("結束Error等級設定")
}

func init_config() {

	//open設定檔(config.json)
	file, err := os.Open("config.json")
	log_info.Info("打開config設定檔")
	//file, err := os.Open("D:\\10_read_daily_clock_in_record_into_mongodb_config\\config.json") //取相對路徑

	buf := make([]byte, 2048)
	if err != nil {
		log_err.WithFields(logrus.Fields{
			"err": err,
		}).Error("打開config錯誤")
	}

	//read 設定檔
	n, err := file.Read(buf)
	fmt.Println(string(buf))
	if err != nil {
		log_err.WithFields(logrus.Fields{
			"err": err,
		}).Error("讀取config錯誤")
		fmt.Println(err)
		panic(err)
	}

	//to json
	log_info.Info("轉換config成json")
	err = json.Unmarshal(buf[:n], &config)
	if err != nil {
		log_err.WithFields(logrus.Fields{
			"err": err,
		}).Error("轉換config成json錯誤")
		fmt.Println(err)
		panic(err)
	}
}

func main() {

	// 建立新路由
	router := mux.NewRouter()

	// 建立路由路徑與Handler
	router.HandleFunc(`/SouthSience/RtspConfig/query`, dailyAPIHandler)

	log_info.WithFields(logrus.Fields{}).Info("建立路由路徑")

	//讀入config
	// readConfig()

	fmt.Println("API port:", config.PortOfAPI)

	// 建立Server
	apiServerPointer := &http.Server{
		Addr:           ":" + config.PortOfAPI,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log_info.WithFields(logrus.Fields{}).Info("建立Server")

	log.Fatal(apiServerPointer.ListenAndServe())

}

func dailyAPIHandler(w http.ResponseWriter, r *http.Request) {

	mongoClientPointer, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(`mongodb://localhost:27017`)) // 連接預設主機

	log_info.WithFields(logrus.Fields{}).Info("連線資料庫")

	//連線錯誤
	if nil != err {
		fmt.Fprintf(w, "連線資料庫錯誤")
		log_err.WithFields(logrus.Fields{"err": err}).Error("連線資料庫錯誤")
		return
	}

	//取Collection
	content, err := mongoClientPointer.
		Database(`SouthernScience`).
		Collection(`rtsp_config`).
		Find(context.TODO(), bson.M{"EnableAudioStream": false}) //find all

	log_info.WithFields(logrus.Fields{}).Info("找collection全部")

	if nil != err {
		fmt.Fprintf(w, "查Collection錯誤") // 寫入回應
		log_err.WithFields(logrus.Fields{"err": err}).Error("查Collection錯誤")
		return
	}

	var results []objects.RtspConfig

	for content.Next(context.TODO()) { // 針對每一紀錄

		var rtspConfig objects.RtspConfig

		err = content.Decode(&rtspConfig) // 解析紀錄

		if nil != err {
			fmt.Fprintf(w, "解析錯誤") // 寫入回應
			log_err.WithFields(logrus.Fields{"err": err}).Error("解析錯誤")

			return
		}

		results = append(results, rtspConfig) // 儲存紀錄

	}

	jsonBytes, err := json.Marshal(results) // 轉成JSON

	if nil != err {
		fmt.Fprintf(w, "轉成JSON錯誤") // 寫入回應
		log_err.WithFields(logrus.Fields{"err": err}).Error("轉成JSON錯誤")
		return
	}

	fmt.Fprintf(w, "%s", string(jsonBytes)) // 寫入回應

}

// 讀設定檔
// func readConfig() {

// 	File, err := os.Open("config.json")
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	fmt.Println("成功開啟 config.json")
// 	defer File.Close()

// 	byteValue, _ := ioutil.ReadAll(File)

// 	//將json設定轉存入變數
// 	err = json.Unmarshal(byteValue, &config)
// 	if err != nil {
// 		panic(err)

// 		log_err.WithFields(logrus.Fields{
// 			"err": err,
// 		}).Error("將設定讀到config變數中失敗")

// 		fmt.Println(err)
// 	}
// }
