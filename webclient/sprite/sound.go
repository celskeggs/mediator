package sprite

type Sound struct {
	File    string `json:"file"`
	Repeat  bool   `json:"repeat"`
	Wait    bool   `json:"wait"`
	Channel uint   `json:"channel"`
	Volume  uint   `json:"volume"`
}
