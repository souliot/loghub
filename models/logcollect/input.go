package logcollect

import (
	"context"
	"strings"
	"time"

	slog "github.com/souliot/siot-log"
)

type Log struct {
	DateTime    time.Time `orm:"pk;column(DateTime);type(datetime)"`
	Date        time.Time `orm:"column(Date);null"`
	Level       string    `orm:"column(Level)"`
	FuncCall    string    `orm:"column(FuncCall)"`
	ServiceName string    `orm:"column(ServiceName)"`
	Address     string    `orm:"column(Address)"`
	Message     string    `orm:"column(Message)"`
}

func (m *Log) TableName() string {
	return "darwin_local"
}

var (
	lineRegex = []string{
		`(?P<DateTime>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\.\d{3}\s*\[(?P<Level>\w{1})\]\s*\[(?P<FuncCall>\S*:\d*)\]\s*(?P<ServiceName>\w*)\s*(?P<Address>\d*\.\d*\.\d*\.\d*:*\d*)\s*(?P<Message>.*)`,
		`(?P<DateTime>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\.\d{3}\s*\[(?P<Level>\w{1})\]\s*\[(?P<FuncCall>\S*:\d*)\]\s*\[(?P<ServiceName>\w*)\]\s*\[(?P<Address>\d*\.\d*\.\d*\.\d*:*\d*)\]\s*(?P<Message>.*)`,
		`(?P<DateTime>\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})\.\d{3}\s*\[(?P<Level>\w{1})\]\s*\[(?P<FuncCall>\S*:\d*)\]\s*(?P<Message>.*)`,
	}
	mlog     = make(chan *Log, 30000)
	mlog_chs chan bool
	todo     = context.TODO()
)

type Input struct {
	lineRegex []string
	paths     []string
	gr        int
}

func NewInput(coll_path []string, gr int) (i *Input) {
	if gr <= 0 {
		gr = 100
	}
	ps := make([]string, 0)
	for _, v := range coll_path {
		ps = append(ps, strings.TrimRight(v, "/")+"/"+"*.log")
	}
	return &Input{
		lineRegex: lineRegex,
		paths:     ps,
		gr:        gr,
	}
}

func (m *Input) Run() {
	var linesChans []chan string
	var err error
	to := TailOptions{
		ReadFrom:  "last",
		Stop:      false,
		Poll:      true,
		StateFile: "logs/",
	}
	tc := Config{
		Paths:   m.paths,
		Type:    RotateStyleSyslog,
		Options: to,
	}
	linesChans, err = GetEntries(todo, tc)
	if err != nil {
		slog.Error("Error occurred while trying to tail logfile")
		return
	}
	rlp, err := NewRegexLineParser(m.lineRegex)
	if err != nil {
		slog.Error("Error occurred while get regex")
		return
	}
	mlog_chs = make(chan bool, m.gr)
	for _, lines := range linesChans {
		go func(plinex chan string, rlp *RegexLineParser) {
			for line := range plinex {
				mlog_chs <- true
				go m.parseLine(line, rlp)
			}
		}(lines, rlp)
	}
}

func (m *Input) Stop() {

}

func (m *Input) parseLine(line string, rlp *RegexLineParser) {
	defer func() {
		<-mlog_chs
	}()
	parsedLine, err := rlp.ParseLine(line)
	if err != nil || len(parsedLine) == 0 {
		return
	}
	log := &Log{}
	datetime, err := time.ParseInLocation("2006/01/02 15:04:05", parsedLine["DateTime"].(string), time.Local)
	if err != nil {
		slog.Error("时间转换错误：", err)
		return
	}

	if sm, ok := parsedLine["ServiceName"]; ok {
		log.ServiceName = sm.(string)
	}
	if addr, ok := parsedLine["Address"]; ok {
		log.Address = addr.(string)
	}
	log.DateTime = datetime
	log.Level = parsedLine["Level"].(string)
	log.FuncCall = parsedLine["FuncCall"].(string)
	log.Message = parsedLine["Message"].(string)
	mlog <- log
}