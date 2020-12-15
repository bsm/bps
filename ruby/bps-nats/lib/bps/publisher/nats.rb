require 'cgi'
require 'nats/io/client'

module BPS
  module Publisher
    class NATS < Abstract
      FLUSH_TIMEOUT = 5

      class Topic < Abstract::Topic
        def initialize(client, topic)
          super()

          @client = client
          @topic = topic
        end

        def publish(message, **_opts)
          @client.publish(@topic, message)
        end

        def flush(**)
          @client.flush(FLUSH_TIMEOUT)
        end
      end

      CLIENT_OPTS = {
        servers: [:string],
        dont_randomize_servers: :bool,
        reconnect_time_wait: :float,
        max_reconnect_attempts: :int,
        connect_timeout: :float,
        # TODO: review, list all of them: https://github.com/nats-io/nats-pure.rb (there's tls config etc)
      }.freeze

      def self.parse_url(url)
        port = url.port&.to_s || '4222'
        servers = CGI.unescape(url.host).split(',').map do |host|
          addr = "nats://#{host}"
          addr << ':' << port unless addr.match(/:\d+$/)
          addr
        end
        opts = CGI.parse(url.query || '').transform_values {|v| v.size == 1 ? v[0] : v }
        opts.merge(servers: servers)
      end

      # @return [BPS::Coercer] the options coercer.
      def self.coercer
        @coercer ||= BPS::Coercer.new(CLIENT_OPTS).freeze
      end

      # @param [Hash] options.
      def initialize(**opts)
        super()

        @topics = {}
        @client = ::NATS::IO::Client.new
        @client.connect(**opts.slice(*CLIENT_OPTS.keys))
      end

      def topic(name)
        @topics[name] ||= self.class::Topic.new(@client, name)
      end

      def close
        @client.close
      end
    end
  end
end
