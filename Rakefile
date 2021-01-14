require 'bundler/setup'
require 'rspec/core/rake_task'
require 'rubocop/rake_task'
require 'bundler/gem_helper'

RSpec::Core::RakeTask.new(:spec) do |t|
  t.pattern = 'ruby/*/spec/**{,/*/**}/*_spec.rb'
end

RuboCop::RakeTask.new(:rubocop)

PACKAGES = Dir['ruby/*/*.gemspec'].map {|fn| File.dirname(fn) }.freeze

namespace :pkg do
  PACKAGES.each do |dir|
    name = File.basename(dir)
    namespace name.to_sym do
      Bundler::GemHelper.install_tasks dir: File.expand_path(dir, __dir__)
    end
  end
end

desc 'Release and publish all'
task release: PACKAGES.map {|dir| "pkg:#{File.basename(dir)}:release" }

task default: %i[spec rubocop]
