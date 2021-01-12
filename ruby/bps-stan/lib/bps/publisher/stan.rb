require 'cgi'
require 'stan/client'

module BPS
  module Publisher
    class STAN < Abstract
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
          # noop
        end
      end

      CLIENT_OPTS = {
        nats: {
          servers: [:string],
          dont_randomize_servers: :bool,
          reconnect_time_wait: :float,
          max_reconnect_attempts: :int,
          connect_timeout: :float,
          tls_ca_file: :string,
          # TODO: review, list all of them: https://github.com/nats-io/nats.rb (there's tls config etc)
        },
      }.freeze

      def self.parse_url(url)
        port = url.port&.to_s || '4222'
        servers = CGI.unescape(url.host).split(',').map do |host|
          addr = "nats://#{host}"
          addr << ':' << port unless addr.match(/:\d+$/)
          addr
        end
        opts = CGI.parse(url.query || '').transform_values {|v| v.size == 1 ? v[0] : v }
        cluster_id = opts.delete('cluster_id')
        client_id = opts.delete('client_id')
        [cluster_id, client_id, { nats: { servers: servers } }]
      end

      # @return [BPS::Coercer] the options coercer.
      def self.coercer
        @coercer ||= BPS::Coercer.new(CLIENT_OPTS).freeze
      end

      # @param [String] cluster ID.
      # @param [String] client ID.
      # @param [Hash] options.
      def initialize(cluster_id, client_id, nats: {}, **opts)
        super()

        # handle TLS if CA file is provided:
        if !nats[:tls] && nats[:tls_ca_file]
          ctx = OpenSSL::SSL::SSLContext.new
          ctx.set_params
          ctx.ca_file = nats.delete(:tls_ca_file)
          nats[:tls] = ctx
        end

        @topics = {}
        @client = ::STAN::Client.new
        @client.connect(cluster_id, client_id, nats: nats, **opts.slice(*CLIENT_OPTS.keys))
      end

      def topic(name)
        @topics[name] ||= self.class::Topic.new(@client, name)
      end

      def close
        # NATS/STAN do not survive multi-closes, so close only once:
        @client&.close
        @client = nil
      end
    end
  end
end
