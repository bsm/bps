require 'bps'
require 'kafka'

module BPS::Kafka
  # ReliablePublisher is a slow, but reliable Kafka Publisher.
  # There is no batching at all, messages are emitted synchronously.
  class ReliablePublisher < BPS::Publisher::Abstract
    DELIVER_MESSAGE_RETRIES = 3 # 1 by kafka-ruby's Kafka::Client default, increase slightly

    class Topic < BPS::Publisher::Topic::Abstract
      def initialize(client, _topic_name)
        @client = client
        @topic_name = @topic_name
      end

      def publish(msg_data)
        @client.deliver_message(msg_data, topic: @topic_name, retries: DELIVER_MESSAGE_RETRIES)
        nil
      end
    end

    def initialize(seed_brokers, client_id: nil)
      @client = Kafka.new(seed_brokers, client_id: client_id)
    end

    def topic(name)
      Topic.new(@client, name)
    end

    def close
      @client.close
      nil
    end
  end
end
