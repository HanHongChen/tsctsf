package processor

import (
	"github.com/HanHongChen/tsctsf/internal/sbi/consumer"
	"github.com/HanHongChen/tsctsf/pkg/app"
)

type TSCTSF interface {
	app.App
	Consumer() *consumer.Consumer
}

type Processor struct {
	TSCTSF
}

func NewProcessor(tsctsf TSCTSF) (*Processor, error) {
	p := &Processor{
		TSCTSF: tsctsf,
	}
	return p, nil
}
