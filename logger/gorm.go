package logger

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type (
	ctxKey int
)

const (
	// color
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Magenta     = "\033[35m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"

	// ctx key
	KeyRecordSqlFn ctxKey = iota
)

type (
	Writer interface {
		Printf(string, ...interface{})
	}
	Config struct {
		SlowThreshold time.Duration
		Colorful      bool
		LogLevel      gormLogger.LogLevel
	}
	SqlRecord struct {
		Sql        string    `json:"sql"`
		Level      string    `json:"level"`
		BeginTime  time.Time `json:"beginTime"`
		Source     string    `json:"source"`
		ExecTime   string    `json:"execTime"`
		AffectRows int64     `json:"affectRows"`
	}
	logger struct {
		Writer
		Config
		infoStr, warnStr, errStr            string
		traceStr, traceErrStr, traceWarnStr string
		IgnoreRecordNotFoundError           bool
	}
)

var (
	GormLogger = NewGormLogger(log.New(log.Writer(), "\n", log.LstdFlags), Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      gormLogger.Warn,
		Colorful:      true,
	})
)

func NewGormLogger(writer Writer, config Config) gormLogger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}

	return &logger{
		Writer:       writer,
		Config:       config,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

// LogMode log mode
func (l *logger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Info {
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l *logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l *logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
//
//nolint:cyclop
func (l *logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	level := "info"
	source := utils.FileWithLineNum()
	sql, rows := fc()
	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormLogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
	if ctx == nil {
		return
	}
	if f, ok := ctx.Value(KeyRecordSqlFn).(func(record SqlRecord)); ok {
		f(SqlRecord{
			Sql:        sql,
			Level:      level,
			BeginTime:  begin,
			Source:     source,
			ExecTime:   elapsed.String(),
			AffectRows: rows,
		})
	}
}

func SetRecordSqlFn(ctx context.Context, fn func(record SqlRecord)) context.Context {
	return context.WithValue(ctx, KeyRecordSqlFn, fn)
}
