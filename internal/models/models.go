package models

import (
	"time"
)

type Checgolded struct {
	ID        int `gorm:"primaryKey"`
	MessageID int
}

type DlrResponse struct {
	ID       int64  `gorm:"primaryKey"`
	Response string `gorm:"type:text"`
	Company  string `gorm:"size:255"`
}


func (DlrResponse) TableName() string {
    return "dlr_response"
}

func (DlrSms) TableName() string {
	return "dlr_sms"
}

func (ErrorLog) TableName() string {
	return "error_log"
}

func (FailedJob) TableName() string {
	return "failed_jobs"
}

type DlrSms struct {
	ID           int64      `gorm:"primaryKey"`
	MessageID    string     `gorm:"size:255"`
	MessageState string     `gorm:"size:150"`
	ErrorCode    string     `gorm:"size:255"`
	MobileNo     int64      `gorm:"type:bigint"`
	CurrentTime  time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	Data         string     `gorm:"type:text"`
}

type ErrorLog struct {
	ID          int64     `gorm:"primaryKey"`
	Data        string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	Status      int8      `gorm:"default:0"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	IsDelete    int8      `gorm:"default:0"`
}

type FailedJob struct {
	ID         uint64    `gorm:"primaryKey"`
	Connection string    `gorm:"type:text"`
	Queue      string    `gorm:"type:text"`
	Payload    string    `gorm:"type:longtext"`
	Exception  string    `gorm:"type:longtext"`
	FailedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UUID       string    `gorm:"size:500"`
}

type Job struct {
	ID          uint64    `gorm:"primaryKey"`
	Queue       string    `gorm:"size:191"`
	Payload     string    `gorm:"type:longtext"`
	Attempts    uint8     `gorm:"type:tinyint unsigned"`
	ReservedAt  *uint32   `gorm:"type:int unsigned"`
	AvailableAt uint32    `gorm:"type:int unsigned"`
	CreatedAt   uint32    `gorm:"type:int unsigned"`
}

type List struct {
	ID        uint64    `gorm:"primaryKey"`
	Name      string    `gorm:"size:191"`
	CreatedAt time.Time `gorm:"default:NULL"`
	UpdatedAt time.Time `gorm:"default:NULL"`
}

type Messagesnotsent struct {
	ID        int64 `gorm:"primaryKey"`
	Mobile    int64 `gorm:"type:bigint"`
	MessageID int64 `gorm:"type:bigint"`
}

type Migration struct {
	ID        uint `gorm:"primaryKey"`
	Migration string `gorm:"size:191"`
	Batch     int
}

type Number struct {
	ID        uint64 `gorm:"primaryKey"`
	Number    int64  `gorm:"type:bigint"`
	ListID    int    `gorm:"type:int"`
	CreatedAt time.Time `gorm:"default:NULL"`
	UpdatedAt time.Time `gorm:"default:NULL"`
}

type Number2 struct {
	ID        uint64 `gorm:"primaryKey"`
	Number    int64  `gorm:"type:bigint"`
	CreatedAt time.Time `gorm:"default:NULL"`
	UpdatedAt time.Time `gorm:"default:NULL"`
}

type NumberHlr struct {
	ID           int64     `gorm:"primaryKey"`
	Number       int64     `gorm:"type:bigint"`
	Responser    string    `gorm:"size:255"`
	NetworkName  string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"default:NULL"`
	IsAvailable  int8      `gorm:"default:0"`
	CheckCount   int       `gorm:"default:1"`
}

type NumberReport struct {
	ID           int64      `gorm:"primaryKey"`
	ReportsID    int64      `gorm:"type:bigint"`
	Number       int64      `gorm:"type:bigint"`
	Status       int8       `gorm:"default:1"`
	MessageID    string     `gorm:"size:100"`
	MessageState string     `gorm:"size:255"`
	ErrorCode    string     `gorm:"size:255"`
	Company      string     `gorm:"size:255"`
	Ported       string     `gorm:"size:255"`
	CountPorted  int16      `gorm:"type:smallint;default:0"`
	Sent         int8       `gorm:"default:0"`
	Message      string     `gorm:"size:2000"`
	IsReplay     int8       `gorm:"default:0"`
	CountSend    int        `gorm:"default:0"`
	CreatedAt    time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	SendAt       time.Time  `gorm:"default:NULL"`
}


type Report struct {
	ID                  uint64    `gorm:"primaryKey"`
	TotalNumbers        int64     `gorm:"type:bigint"`
	FilterNumbers       int64     `gorm:"type:bigint"`
	Sender              string    `gorm:"type:varchar(191) binary"`
	Point               int       `gorm:"default:1"`
	RepeatNumber        int64     `gorm:"type:bigint"`
	Message             string    `gorm:"type:varchar(2000) binary"`
	CreatedAt           time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time `gorm:"default:NULL"`
	MessageID           int64     `gorm:"type:bigint"`
	Status              int8      `gorm:"default:0;comment:'{-1=Faild some jobs,0=Pending, 1=Under Processing, 2=Sent}'"`
	DispatchJob         int8      `gorm:"default:0"`
	DispatchJobFailed   int8      `gorm:"default:0"`
	IsVariables         int8      `gorm:"default:0"`
	StcJobCount         int       `gorm:"default:0"`
	ZainJobCount        int       `gorm:"default:0"`
	MobilyJobCount      int       `gorm:"default:0"`
	StcJobFinishedAt    time.Time `gorm:"default:NULL"`
	ZainJobFinishedAt   time.Time `gorm:"default:NULL"`
	MobilyJobFinishedAt time.Time `gorm:"default:NULL"`
	PortedCount         int       `gorm:"default:0"`
}

type ShortUrl struct {
	ID             int64     `gorm:"primaryKey"`
	Uniqid         string    `gorm:"size:250"`
	ShortIdentifier string `gorm:"type:text"`
	LongUrl        string `gorm:"type:text"`
	CreatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Test struct {
	Number  int64     `gorm:"primaryKey;type:bigint"`
	Date    time.Time `gorm:"primaryKey"`
	Message string    `gorm:"type:text"`
}

type User struct {
	ID                uint64    `gorm:"primaryKey"`
	Name              string    `gorm:"size:191"`
	Email             string    `gorm:"size:191;unique"`
	EmailVerifiedAt   time.Time `gorm:"default:NULL"`
	Password          string    `gorm:"size:191"`
	RememberToken     string    `gorm:"size:100"`
	CreatedAt         time.Time `gorm:"default:NULL"`
	UpdatedAt         time.Time `gorm:"default:NULL"`
}

type WhitelistIP struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int       `gorm:"type:int"`
	IP        string    `gorm:"size:250"`
	Name      string    `gorm:"size:500"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type WhiteListNumber struct {
	ID        uint64 `gorm:"primaryKey"`
	Number    int64  `gorm:"type:bigint"`
	CreatedAt time.Time `gorm:"default:NULL"`
	UpdatedAt time.Time `gorm:"default:NULL"`
}