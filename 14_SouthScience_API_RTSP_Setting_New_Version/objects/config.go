package objects

// 設定檔
// type Config struct {
// 	PortOfAPI string //API port
// }

type Config struct {
	Local struct {
	} `json:"local"`

	Databases struct {
		MongoDB struct {
			Server     string `json:"server"`
			Port       string `json:"port"`
			Userid     string `json:"userid"`
			Password   string `json:"password"`
			Database   string `json:"database"`
			Collection string `json:"collection"`
		} `json:"MongoDB"`
	} `json:"databases"`

	Servers struct {
		RTSPSettingServer struct {
			Host string `json:"host"`
			Path string `json:"path"`
			Port string `json:"port"`
		} `json:"RTSPSettingServer"`
	} `json:"servers"`
}
