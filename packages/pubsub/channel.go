package pubsub

func hey() {
	//TODO

	//// Requires only publishConfig
	//func PublishChannelRequest[SubscribeAddressType,
	//	SubscribeDataType,
	//	PublishAddressType,
	//	PublishDataType MemoryArchitecture](processor ChannelProcessor[SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType]) {
	//
	//	publishEvent := ReadEvent[SubscribeAddressType, SubscribeDataType]{
	//		EventType: Event{
	//			Operation: processor.getChannelConfig().PublishType,
	//			Exchange:  Response,
	//		},
	//		AddressType: processor.getChannelConfig().publishData.Data(),
	//		DataType:    processor.getChannelConfig().publishData.Address(),
	//	}
	//
	//	PbManager.Publish(publishEvent.Event(), nil)
	//}
	//
	//func SubscribeChannelRequest[SubscribeAddressType,
	//	SubscribeDataType,
	//	PublishAddressType,
	//	PublishDataType MemoryArchitecture](processor ChannelProcessor[SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType]) chan EventChannelBase {
	//
	//	return PbManager.Subscribe(Event{
	//		Operation: processor.getChannelConfig().SubscribeType,
	//		Exchange:  Request,
	//	})
}
