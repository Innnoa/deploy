package common

type OAServer struct {
	ID         string `json:"id"`
	ServerName string `json:"serverName"`
	IP         string `json:"ip"`
	UserName   string `json:"username"`
	Password   string `json:"password"`
	RootPath   string `json:"rootPath"`
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

type PrinterWithPackage struct {
	ID        string `json:"id"`
	PolNo     string `json:"pol"`
	IP        string `json:"ip"`
	AppId     string `json:"appid"`
	AppName   string `json:"appname"`
	BrandId   string `json:"brandId"`
	AppType   string `json:"apptype"`
	Path      string `json:"installpath"`
	WinFile   string `json:"winfile"`
	UOSFile   string `json:"uosdeb"`
	KylinFile string `json:"kylindeb"`
}

type Printer struct {
	ID    string `json:"id"`
	PolNo string `json:"pol"`
	IP    string `json:"ip"`
	AppId string `json:"appid"`
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
	PolNo     string `json:"pol"`
	IP        string `json:"ip"`
	Reboot    string `json:"reboot"`
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

type AppId struct {
	ID string `json:"appid"`
}

type InstallInfo struct {
	Pols   []string `json:"pols"`
	AppIds []string `json:"appids"`
}

type DetailComputerInfo struct {
	PolNo        string `json:"pol"`
	IP           string `json:"ip"`
	OS           string `json:"os"`
	SP           string `json:"sp"`
	Seedlabel    string `json:"seedlabel"`
	SystemDrive  string `json:"systemdrive"`
	NumOfDrive   string `json:"numofdrive"`
	LastDrive    string `json:"lastdrive"`
	SizeOfDrive1 string `json:"sizeofdrive1"`
	SizeOfDrive2 string `json:"sizeofdrive2"`
	FreeSpaceC   string `json:"freespacec"`
	FreeSpaceD   string `json:"freespaced"`
	CpuSpeed     string `json:"cpuspeed"`
	CpuType      string `json:"cputype"`
	Ram          string `json:"ram"`
	PCModel      string `json:"pcmodel"`
	BootEnv      string `json:"bootenv"`
	LastSignon   string `json:"lastsignon"`
	LogonId      string `json:"logonid"`
	KBCode       string `json:"kbcode"`
}

type SeedLabelInfo struct {
	Id        string `json:"id"`
	SeedLabel string `json:"seedlabel"`
	Status    string `json:"status"`
}

type TempInfo struct {
	Packages []PackageInfo `json:"packages"`
	Server   OAServer      `json:"server"`
	Computer ComputerInfo  `json:"computer"`
}

type AppVersionInfo struct {
	Version      string `json:"version"`
	Type         string `json:"type"`
	DownloadUrl  string `json:"downloadUrl"`
	ReleaseNotes string `json:"releaseNotes"`
}

type SeedTimeInfo struct {
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	SeedLabel  string `json:"seedlabel"`
}
