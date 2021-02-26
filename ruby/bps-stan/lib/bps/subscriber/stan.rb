require 'cgi'
require 'stan/client'

module BPS
  module Subscriber
    class STAN < Abstract
      def initialize(cluster_id, client_id, nats: {}, **opts)
        super()

        # handle TLS if CA file is provided:
        if !nats[:tls] && nats[:tls_ca_file]
          ctx = OpenSSL::SSL::SSLContext.new
          ctx.set_params
          ctx.ca_file = nats.delete(:tls_ca_file)
          nats[:tls] = ctx
        end

        @client = ::STAN::Client.new
        @client.connect(cluster_id, client_id, nats: nats, **opts.slice(*::BPS::STAN::Utils::CLIENT_OPTS.keys))
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
        # NATS/STAN does not survive multi-closes, so close only once:
        @client&.close
        @client = nil
      end
    end
  end
end
