require 'bps/stan'
require 'spec_helper'

NATS_CLUSTER = 'test-cluster'.freeze
NATS_CLIENT  = 'bps-test'.freeze

def nats_servers
  ENV.fetch('NATS_SERVERS', '127.0.0.1:4222').split(',').freeze
end

def nats_servers_with_scheme
  nats_servers.map {|s| "nats://#{s}" }
end

run_spec = \
  begin
    # passing a block means that client will be automatically closed once block exits:
    ::STAN::Client.new.connect(NATS_CLUSTER, NATS_CLUSTER, nats: { servers: nats_servers_with_scheme, dont_randomize_servers: true }) {}
    true
  rescue StandardError => e
    warn "WARNING: unable to run #{File.basename __FILE__}: #{e.message}"
    false
  end

helper = proc do
  def read_messages(topic_name, num_messages)
    [].tap do |messages|
      ::STAN::Client.new.connect(NATS_CLUSTER, NATS_CLUSTER, nats: { servers: nats_servers_with_scheme, dont_randomize_servers: true }) do |client|
        client.subscribe(topic_name, start_at: :first) do |msg|
          messages << msg.data
          next if messages.size == num_messages
        end
      end
    end
  end
end

RSpec.describe 'STAN' do
  context 'resolve addrs' do
    let(:publisher) { double('BPS::Publisher::STAN') }
    before          { allow(BPS::Publisher::STAN).to receive(:new).and_return(publisher) }

    it 'should resolve simple URLs' do
      expect(BPS::Publisher::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://test.host:4222'] })
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('stan://test.host:4222?cluster_id=CLUSTER&client_id=CLIENT'))
    end

    it 'should resolve URLs with multiple hosts' do
      expect(BPS::Publisher::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://foo.host:4222', 'nats://bar.host:4222'] })
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('stan://foo.host,bar.host:4222?cluster_id=CLUSTER&client_id=CLIENT'))
    end

    it 'should resolve URLs with multiple hosts/ports' do
      expect(BPS::Publisher::STAN)
        .to receive(:new)
        .with('CLUSTER', 'CLIENT', nats: { servers: ['nats://foo.host:4223', 'nats://bar.host:4222'] })
        .and_return(publisher)
      BPS::Publisher.resolve(URI.parse('stan://foo.host%3A4223,bar.host?cluster_id=CLUSTER&client_id=CLIENT'))
    end
  end

  context BPS::Publisher::STAN, if: run_spec do
    it_behaves_like 'publisher', url: "stan://#{CGI.escape(nats_servers.join(','))}/?cluster_id=#{NATS_CLUSTER}&client_id=#{NATS_CLIENT}", &helper
  end
end