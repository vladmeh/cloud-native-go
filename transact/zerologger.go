package transact

import "cloudNativeGo/core"

type ZeroTransactionLogger struct{}

func (z *ZeroTransactionLogger) WriteDelete(key string)                        {}
func (z *ZeroTransactionLogger) WritePut(key, value string)                    {}
func (z *ZeroTransactionLogger) Err() <-chan error                             { return nil }
func (z *ZeroTransactionLogger) LastSequence() uint64                          { return 0 }
func (z *ZeroTransactionLogger) Run()                                          {}
func (z *ZeroTransactionLogger) Wait()                                         {}
func (z *ZeroTransactionLogger) Close() error                                  { return nil }
func (z *ZeroTransactionLogger) ReadEvents() (<-chan core.Event, <-chan error) { return nil, nil }
func NewZeroTransactionLogger() (core.TransactionLogger, error) {
	return &ZeroTransactionLogger{}, nil
}
