package pubsub

type ChannelConfig[
	SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType MemoryArchitecture] struct {
	SubscribeType OperationType
	subscribeData ReadEvent[SubscribeAddressType, SubscribeDataType]
	PublishType   OperationType
	publishData   ReadEvent[PublishAddressType, PublishDataType]
}

type ChannelProcessor[SubscribeAddressType,
	SubscribeDataType,
	PublishAddressType,
	PublishDataType MemoryArchitecture] interface {
	processReadEvent(SubscribeAddressType) SubscribeDataType
	processWriteEvent(PublishAddressType, PublishDataType)
	getChannelConfig() ChannelConfig[SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType]
}

// Requires only publishConfig
func PublishChannelRequest[SubscribeAddressType,
	SubscribeDataType,
	PublishAddressType,
	PublishDataType MemoryArchitecture](processor ChannelProcessor[SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType]) {

	publishEvent := ReadEvent[SubscribeAddressType, SubscribeDataType]{
		EventType: Event{
			Operation: processor.getChannelConfig().PublishType,
			Exchange:  Response,
		},
		AddressType: processor.getChannelConfig().publishData.Data(),
		DataType:    processor.getChannelConfig().publishData.Address(),
	}

	PbManager.Publish(publishEvent.Event(), nil)
}

func SubscribeChannelRequest[SubscribeAddressType,
	SubscribeDataType,
	PublishAddressType,
	PublishDataType MemoryArchitecture](processor ChannelProcessor[SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType]) chan EventChannelBase {

	return PbManager.Subscribe(Event{
		Operation: processor.getChannelConfig().SubscribeType,
		Exchange:  Request,
	})
}

// Requires pubsub channel config
func ProcessChannelTransactions[SubscribeAddressType,
	SubscribeDataType,
	PublishAddressType,
	PublishDataType MemoryArchitecture](
	processor ChannelProcessor[SubscribeAddressType, SubscribeDataType, PublishAddressType, PublishDataType]) {
	subscribeChannel := PbManager.Subscribe(Event{
		Operation: processor.getChannelConfig().SubscribeType,
		Exchange:  Request,
	})

	//Subscribe loop

	// Subscribe to (EventChannel)

	//Handle data in proccessSubscribe if there is a config for this MemoryRead

	//If there is a publish config then do the following

	//Process data if there is a publishFunction

	//Respond in Publish if there is a publish EventChannel

	//Publish Loop

	//Publish to EventChannel

	//Scenario

	//Get ram adress

	//Publish Re

	for {
		select {
		case subEvent := <-subscribeChannel:

			// Process ReadData
			data2 := processor.processReadEvent(subEvent.Data)

			responseEvent := ReadEvent[SubscribeAddressType, SubscribeDataType]{
				EventType: Event{
					Operation: processor.getChannelConfig().PublishType,
					Exchange:  Response,
				},
				AddressType: processor.getChannelConfig().subscribeData.AddressType,
				DataType:    data2,
			}
			PbManager.Publish(responseEvent.Event(), responseEvent)

			/*		writeChannel := PbManager.Publish(Event{
					Operation: processor.getChannelConfig().SubscribeType,
					Exchange:  Request,
				})*/

			// Process WriteData
			PublishChannelRequest(processor)
		}

	}
}
