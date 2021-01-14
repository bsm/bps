require 'spec_helper'
require 'bps/kafka'

RSpec.describe 'Kafka', kafka: true do
  context 'with addr resolving' do
    let(:client) { instance_double(::Kafka::Client, producer: nil) }

    before       { allow(::Kafka).to receive(:new).and_return(client) }

    it 'resolves simple URLs' do
      if Gem::Version.new(RUBY_VERSION) < Gem::Version.new('2.7')
        allow(::Kafka).to receive(:new).with(['test.host:9092'], {}).and_return(client)
      else
        allow(::Kafka).to receive(:new).with(['test.host:9092']).and_return(client)
      end
      BPS::Publisher.resolve(URI.parse('kafka+sync://test.host:9092'))
    end

    it 'resolves URLs with multiple hosts' do
      if Gem::Version.new(RUBY_VERSION) < Gem::Version.new('2.7')
        allow(::Kafka).to receive(:new).with(['foo.host:9092', 'bar.host:9092'], {}).and_return(client)
      else
        allow(::Kafka).to receive(:new).with(['foo.host:9092', 'bar.host:9092']).and_return(client)
      end
      BPS::Publisher.resolve(URI.parse('kafka+sync://foo.host,bar.host:9092'))
    end

    it 'resolves URLs with multiple hosts/ports' do
      if Gem::Version.new(RUBY_VERSION) < Gem::Version.new('2.7')
        allow(::Kafka).to receive(:new).with(['foo.host:9093', 'bar.host:9092'], {}).and_return(client)
      else
        allow(::Kafka).to receive(:new).with(['foo.host:9093', 'bar.host:9092']).and_return(client)
      end
      BPS::Publisher.resolve(URI.parse('kafka+sync://foo.host%3A9093,bar.host'))
    end
  end

  describe 'publishers' do
    let(:kafka_addrs) { ENV.fetch('KAFKA_ADDRS', '127.0.0.1:9092').split(',') }

    def read_messages(topic_name, num_messages)
      client = ::Kafka.new(kafka_addrs)
      Enumerator.new do |y|
        client.each_message(topic: topic_name, start_from_beginning: true) do |msg|
          y << msg.value
        end
      end.take(num_messages)
    ensure
      client&.close
    end

    context BPS::Publisher::Kafka do
      let(:publisher_url) { "kafka+sync://#{CGI.escape(kafka_addrs.join(','))}/" }

      it_behaves_like 'publisher'
    end

    context BPS::Publisher::KafkaAsync do
      let(:publisher_url) { "kafka://#{CGI.escape(kafka_addrs.join(','))}/" }

      it_behaves_like 'publisher'
    end
  end
end
