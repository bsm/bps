require 'bps'
require 'bps/publisher/stan'

module BPS
  module Publisher
    register('stan') do |url, **opts|
      cluster_id, client_id, url_opts = STAN.parse_url(url)
      STAN.new(cluster_id, client_id, **STAN.coercer.coerce(url_opts.merge(opts)))
    end
  end
end
