package processor

import (
	"github.com/tencentmusic/evhub/internal/processor/connector"
	"github.com/tencentmusic/evhub/internal/processor/options"
	"github.com/tencentmusic/evhub/internal/processor/protocol"
	cc "github.com/tencentmusic/evhub/pkg/conf_connector"
	"github.com/tencentmusic/evhub/pkg/log"
)

// Processor is a server of the event
type Processor struct {
	// Protocol is the name of the processor protocol
	Protocol *protocol.Protocol
	// Connector is the name of the processor connector
	Connector connector.Connector
	// Opts is the name of the processor options
	Opts *options.Options
	// ProducerConfConnector is the name of the producer configuration
	ProducerConfConnector cc.ProducerConfConnector
	// ProducerConnector is the name of the processor configuration
	ProcessorConfConnector cc.ProcessorConfConnector
}

// New creates a new Processor
func New(opts *options.Options) (*Processor, error) {
	// initialize producer configuration
	producerConfConnector, err := cc.NewProducerConfConnector(opts.ConfConnectorConfig)
	if err != nil {
		return nil, err
	}
	// initialize processor configuration
	processorConfConnector, err := cc.NewProcessorConfConnector(opts.ConfConnectorConfig)
	if err != nil {
		return nil, err
	}
	// create protocol
	protocol, err := protocol.New(opts)
	if err != nil {
		return nil, err
	}
	// create connector
	c, err := connector.New(opts, &opts.ConnectorConfig, protocol.Handler)
	if err != nil {
		return nil, err
	}

	return &Processor{
		Protocol:               protocol,
		Connector:              c,
		ProcessorConfConnector: processorConfConnector,
		ProducerConfConnector:  producerConfConnector,
	}, nil
}

// Start initialize some operations
func (s *Processor) Start() error {
	producerConfInfoList, err := s.ProducerConfConnector.GetAllConfInfo()
	if err != nil {
		return err
	}
	processorConfInfoList, err := s.ProcessorConfConnector.GetAllConfInfo()
	if err != nil {
		return err
	}
	s.Connector.ConsumerDelayTopic(producerConfInfoList)
	s.Connector.ConsumerProcessorTopic(processorConfInfoList)
	// watch
	if err := s.ProducerConfConnector.SetWatch(s.Connector.WatchDelayTopic); err != nil {
		return err
	}
	// watch
	if err := s.ProcessorConfConnector.SetWatch(s.Connector.WatchProcessorTopic); err != nil {
		return err
	}
	return nil
}

// Stop stops the processor gracefully.
func (s *Processor) Stop() error {
	if err := s.Connector.Stop(); err != nil {
		log.Errorf("connector stop failed err: %v", err)
	}
	if err := s.Protocol.Stop(); err != nil {
		log.Errorf("protocol stop failed err: %v", err)
	}
	return nil
}
