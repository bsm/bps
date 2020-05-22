Gem::Specification.new do |s|
  s.name        = 'bps-kafka'
  s.version     = File.read(File.expand_path('../.version', __dir__)).strip
  s.platform    = Gem::Platform::RUBY

  s.licenses    = ['Apache-2.0']
  s.summary     = 'Kafka adapter for bps'
  s.description = 'https://github.com/bsm/bps'

  s.authors     = ['Dimitrij Denissenko']
  s.email       = 'dimitrij@blacksquaremedia.com'
  s.homepage    = 'https://github.com/bsm/bps'

  s.executables   = []
  s.files         = `git ls-files`.split("\n")
  s.test_files    = `git ls-files -- spec/*`.split("\n")
  s.require_paths = ['lib']
  s.required_ruby_version = '>= 2.7.1'

  s.add_dependency 'bps', s.version
  s.add_dependency 'ruby-kafka', '1.1.0.beta1'
end
