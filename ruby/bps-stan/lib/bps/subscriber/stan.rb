require 'cgi'
require 'stan/client'

module BPS
  module Subscriber
    class STAN < Abstract
      # @param [String] cluster ID.
      # @param [String] client ID.
      # @param [Hash] options.
      def initialize(cluster_id, client_id, **opts)
        super()

        @client = ::BPS::STAN.connect(cluster_id, client_id, **opts)
      end

      # Subscribe to a topic
      # @param topic [String] topic the topic name.
      def subscribe(topic, **opts)
        # important opts:
        # - queue: 'queue-name'          # https://docs.nats.io/developing-with-nats-streaming/queues
        # - durable_name: 'durable-name' # https://docs.nats.io/developing-with-nats-streaming/durables
        @client.subscribe(topic, **opts) do |msg|
          yield msg.data # TODO: maybe yielding just bytes is not too flexible? But IMO can wait till (much) later
        end
      end

      # Close the subscriber.
      def close
        super

        # NATS/STAN does not survive multi-closes, so close only once:
        @client&.close
        @client = nil
      end
    end
  end
end
