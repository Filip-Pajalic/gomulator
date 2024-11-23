package pubsub

import "log"

func ProcessChannelTransactions[AddressType MemoryArchitecture, DataType MemoryArchitecture](
	processor ChannelProcessor[AddressType, DataType],
) {
	// Subscribe to read requests

	readChannel := PbManager.Subscribe(Event{
		Operation: processor.getChannelConfig().ReadType,
		Exchange:  Request,
	})
	writeChannel := PbManager.Subscribe(Event{
		Operation: processor.getChannelConfig().WriteType,
		Exchange:  Request,
	})

	for {
		select {
		case eventBase, ok := <-readChannel:
			if !ok {
				log.Printf("pubsub: Read channel for %s closed.", processor.getChannelConfig().ReadType)
				return // Channel closed
			}
			address, ok := eventBase.Address.(uint16) // Adjust based on AddressType
			if !ok {
				log.Println("pubsub: Invalid AddressType in read event.")
				continue
			}
			responseChan, ok := eventBase.ResponseChan.(chan byte) // Adjust based on DataType
			if !ok && eventBase.ResponseChan != nil {
				log.Println("pubsub: Invalid ResponseChan type in read event.")
				continue
			}

			// Process the read event
			data := processor.processReadEvent(address)

			// Send the data back via the response channel if it exists
			if responseChan != nil {
				select {
				case responseChan <- data:
					// Successfully sent
				default:
					// Response channel is full; handle accordingly
					log.Println("pubsub: Response channel is full. Dropping response.")
				}
			}

		case eventBase, ok := <-writeChannel:
			if !ok {
				log.Printf("pubsub: Write channel for %s closed.", processor.getChannelConfig().WriteType)
				return // Channel closed
			}
			address, ok := eventBase.Address.(uint16) // Adjust based on AddressType
			if !ok {
				log.Println("pubsub: Invalid AddressType in write event.")
				continue
			}
			data, ok := eventBase.Data.(byte) // Adjust based on DataType
			if !ok {
				log.Println("pubsub: Invalid DataType in write event.")
				continue
			}

			// Process the write event
			processor.processWriteEvent(address, data)
		}
	}
}
