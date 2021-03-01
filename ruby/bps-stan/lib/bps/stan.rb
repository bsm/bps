require 'bps'
require 'bps/publisher/stan'
require 'bps/subscriber/stan'

module BPS
  module Publisher
    register('stan') do |url, **opts|
      cluster_id, client_id, url_opts = ::BPS::STAN.parse_url(url)
      url_opts.update(opts)
      STAN.new(cluster_id, client_id, **::BPS::STAN.coercer.coerce(url_opts))
    end
  end

  module Subscriber
    register('stan') do |url, **opts|
      cluster_id, client_id, url_opts = ::BPS::STAN.parse_url(url)
      url_opts.update(opts)
      STAN.new(cluster_id, client_id, **::BPS::STAN.coercer.coerce(url_opts))
    end
  end

  module STAN
    CLIENT_OPTS = {
      nats: {
        servers: [:string],
        dont_randomize_servers: :bool,
        reconnect_time_wait: :float,
        max_reconnect_attempts: :int,
        connect_timeout: :float,
        tls_ca_file: :string,
        # TODO: review, list all of them: https://github.com/nats-io/nats.rb
      },
    }.freeze

    # @param [String] cluster ID
    # @param [String] client ID
    # @param [Hash] options
    # @return [STAN::Client] connected STAN client
    def self.connect(cluster_id, client_id, nats: {}, **opts)
      # handle TLS if CA file is provided:
      if !nats[:tls] && nats[:tls_ca_file]
        ctx = OpenSSL::SSL::SSLContext.new
        ctx.set_params
        ctx.ca_file = nats.delete(:tls_ca_file)
        nats[:tls] = ctx
      end

      client = ::STAN::Client.new
      client.connect(cluster_id, client_id, nats: nats, **opts.slice(*CLIENT_OPTS.keys))
      client
    end

    # @return [BPS::Coercer] the options coercer
    def self.coercer
      @coercer ||= BPS::Coercer.new(CLIENT_OPTS).freeze
    end

    # @return [Array] arguments for connecting to STAN
    def self.parse_url(url)
      port = url.port&.to_s || '4222'
      servers = CGI.unescape(url.host).split(',').map do |host|
        addr = "nats://#{host}"
        addr << ':' << port unless /:\d+$/.match?(addr)
        addr
      end
      opts = CGI.parse(url.query || '').transform_values {|v| v.size == 1 ? v[0] : v }
      cluster_id = opts.delete('cluster_id')
      client_id = opts.delete('client_id')
      [cluster_id, client_id, { nats: { servers: servers } }]
    end
  end
end
