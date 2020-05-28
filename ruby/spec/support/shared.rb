require 'securerandom'

RSpec.shared_examples 'publisher' do |features|
  def read_messages(_topic_name, _num_messages)
    raise 'must be overridden'
  end

  before do
    expect(features).to have_key(:url) # each publisher must register under at least 1 scheme
  end

  subject do
    BFS.resolve_publisher(features[:url])
  end

  after do
    subject.close
  end

  it 'should register' do
    expect(subject).to be_a(described_class)
  end

  it 'should publish' do
    topic_name = "bps-test-topic-#{SecureRandom.uuid}"
    seed_messages = 3.times.map { "bps-test-message-#{SecureRandom.uuid}" }

    topic = subject.topic(topic_name)
    seed_messages.each {|msg| topic.publish(msg) }

    published_messages = features[:read_messages].call(topic_name)
    expect(published_messages).to match_array(seed_messages, seed_messages.count)
  end
end
