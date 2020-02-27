package main

import (
	"github.com/buexplain/go-flog"
	"github.com/buexplain/go-flog/extra"
	"github.com/buexplain/go-flog/formatter"
	"github.com/buexplain/go-flog/handler"
)

var logger *flog.Logger

func init()  {
	h := handler.NewSTD(flog.LEVEL_DEBUG, formatter.NewLine(), flog.LEVEL_ERROR)
	logger = flog.New("test", h, extra.NewFuncCaller(3))
	fh := handler.NewFile(flog.LEVEL_DEBUG, formatter.NewLine(), "./example")
	fh.SetMaxSize(239)
	logger.PushHandler(fh)
}

func main() {
	logger.Info("test info", "more info")
	logger.Error("test error", "more info")
}
