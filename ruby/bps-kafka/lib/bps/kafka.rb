require 'bps'
require 'kafka'

module BPS
  module Kafka
    # ReliablePublisher is a slow, but reliable Kafka Publisher.
    # There is no batching at all, messages are delivered one by one.
    class ReliablePublisher < BPS::Publisher::Abstract
      class Topic < BPS::Publisher::Topic::Abstract
        def initialize(producer, topic_name)
          @producer = producer
          @opts = { topic: topic_name }
        end

        def publish(msg_data, **opts)
          @producer.produce(msg_data, **@opts, **opts)
          @producer.deliver_messages
          nil
        end

        # meant for internal use
        # TODO: make a part of topic "interface"?
        def _close
          @producer.shutdown
        end
      end
      private_constant :Topic

      def initialize(seed_brokers, **opts)
        @client = ::Kafka.new(seed_brokers, **opts)
        @topics = {}
      end

      def topic(name)
        @topics[name] ||= Topic.new(@client.producer, name)
      end

      def close
        @topics.values.each(&:_close)
        @client.close
        nil
      end
    end
  end

  register_publisher('kafka') do |url, opts|
    addrs = CGI.unescape(url.host).split(',')
    Kafka::ReliablePublisher.new(addrs, **opts)
  end
end
