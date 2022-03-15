Gem::Specification.new do |s|
  s.name        = 'bps'
  s.version     = File.read(File.expand_path('../../.version', __dir__)).strip
  s.platform    = Gem::Platform::RUBY

  s.licenses    = ['Apache-2.0']
  s.summary     = 'Multi-platform pubsub adapter'
  s.description = 'Minimalist abstraction for publish-subscribe'

  s.authors     = ['Black Square Media']
  s.email       = 'info@blacksquaremedia.com'
  s.homepage    = 'https://github.com/bsm/bps'

  s.executables   = []
  s.files         = `git ls-files`.split("\n")
  s.test_files    = `git ls-files -- spec/*`.split("\n")
  s.require_paths = ['lib']
  s.required_ruby_version = '>= 2.7'
  s.metadata['rubygems_mfa_required'] = 'true'
end
