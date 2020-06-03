require 'uri'
require 'cgi'

module BPS
  autoload :Coercer, 'bps/coercer'

  module Publisher
    autoload :Abstract, 'bps/publisher/abstract'

    def self.register(*schemes, &resolver)
      @registry ||= {}
      schemes.each do |scheme|
        @registry[scheme] = resolver
      end
    end

    def self.resolve(url, coercer: nil)
      url = url.is_a?(::URI) ? url.dup : URI.parse(url)
      rsl = @registry[url.scheme]
      raise ArgumentError, "Unable to resolve publisher #{url}, scheme #{url.scheme} is not registered" unless rsl

      opts = {}
      CGI.parse(url.query.to_s).each do |key, values|
        opts[key.to_sym] = values.first
      end
      opts = coercer.coercer(opts) if coercer
      rsl.call(url, opts)
    end
  end

  module Subscriber
    autoload :Abstract, 'bps/subscriber/abstract'

    def self.register(*schemes, &resolver)
      @registry ||= {}
      schemes.each do |scheme|
        @registry[scheme] = resolver
      end
    end

    def self.resolve(url, coercer: nil)
      url = url.is_a?(::URI) ? url.dup : URI.parse(url)
      rsl = @registry[url.scheme]
      raise ArgumentError, "Unable to resolve subscriber #{url}, scheme #{url.scheme} is not registered" unless rsl

      opts = {}
      CGI.parse(url.query.to_s).each do |key, values|
        opts[key.to_sym] = values.first
      end
      opts = coercer.coercer(opts) if coercer
      rsl.call(url, opts)
    end
  end
end
