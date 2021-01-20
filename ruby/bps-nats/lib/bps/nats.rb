require 'bps'
require 'bps/publisher/nats'

module BPS
  module Publisher
    register('nats') do |url, **opts|
      url_opts = NATS.parse_url(url)
      url_opts.update(opts)
      NATS.new(**NATS.coercer.coerce(url_opts))
    end
  end
end
