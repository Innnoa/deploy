package common

type OAServer struct {
	ID         string `json:"id"`
	ServerName string `json:"serverName"`
}

type PrinterModel struct {
	ID    string `json:"id"`
	Brand string `json:"brand"`
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
	AppName   string `json:"appname"`
	BrandId   string `json:"brandId"`
	AppType   string `json:"apptype"`
	Path      string `json:"installpath"`
	WinFile   string `json:"winfile"`
	UOSFile   string `json:"uosdeb"`
	KylinFile string `json:"kylindeb"`
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
