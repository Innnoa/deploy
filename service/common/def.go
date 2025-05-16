package common

type OAServer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ComputerInfo struct {
	Name string `json:"name"`
	Seed string `json:"seed"`
	IP   string `json:"ip"`
	OA   string `json:"oa"`
}

type Printer struct {
	ID    string `json:"id"`
	PolNo string `json:"pol"`
	IP    string `json:"ip"`
}

type PackageInfo struct {
	ID        string `json:"id"`
	AppName   string `json:"app_name"`
	AppType   string `json:"app_type"`
	Path      string `json:"path"`
	WinFile   string `json:"win_file"`
	UOSFile   string `json:"uos_file"`
	KylinFile string `json:"kylin_file"`
	Status    string `json:"status"`
	Error     string `json:"error"`
}

type Status int

const (
	Waiting Status = iota // 显式指定类型
	Running
	Completed
	Failed
)

func (s Status) String() string {
	return [...]string{
		"Waiting", "Running", "Completed", "Failed",
	}[s]
}
