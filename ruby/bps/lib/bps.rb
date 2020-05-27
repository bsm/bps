require 'uri'
require 'cgi'

module BPS
  module Publisher
    class Abstract
      # def initialize; end

      def topic(_name)
        raise 'not implemented'
      end

      def close; end
    end

    module Topic
      class Abstract
        # def initialize; end

        def publish(_msg_data)
          raise 'not implemented'
        end
      end
    end
  end

  module Subscriber
    class Abstract
      def initialize; end

      def subscribe(_topic, **_opts)
        raise 'not implemented'
      end

      def close; end
    end
  end

  def self.register_publisher(*schemes, &resolver)
    @publishers ||= {}
    schemes.each do |scheme|
      @publishers[scheme] = resolver
    end
  end

  def self.register_subscriber(*schemes, &resolver)
    @subscribers ||= {}
    schemes.each do |scheme|
      @subscribers[scheme] = resolver
    end
  end

  def self.resolve_publisher(url)
    url = url.is_a?(::URI) ? url.dup : URI.parse(url)
    rsl = @publishers[url.scheme]
    raise ArgumentError, "Unable to resolve publisher #{url}, scheme #{url.scheme} is not registered" unless rsl

    opts = {}
    CGI.parse(url.query.to_s).each do |key, values|
      opts[key.to_sym] = values.first
    end
    rsl.call(url, opts)
  end

  def self.resolve_subscriber(url)
    url = url.is_a?(::URI) ? url.dup : URI.parse(url)
    rsl = @subscribers[url.scheme]
    raise ArgumentError, "Unable to resolve subscriber #{url}, scheme #{url.scheme} is not registered" unless rsl

    opts = {}
    CGI.parse(url.query.to_s).each do |key, values|
      opts[key.to_sym] = values.first
    end
    rsl.call(url, opts)
  end
end
