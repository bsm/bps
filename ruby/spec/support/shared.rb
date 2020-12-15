require 'securerandom'

RSpec.shared_examples 'publisher' do
  # WARNING: This example group requires the following helpers to be defined by caller:
  #   - `read_messages(topic_name, num_messages)`
  #   - `publisher_url`

  subject do
    BPS::Publisher.resolve(publisher_url)
  end

  after do
    subject.close
  end

  it 'should register' do
    expect(subject).to be_a(described_class)
  end

  it 'should publish' do
    topic_name = "bps-test-topic-#{SecureRandom.uuid}"
    messages = 3.times.map { "bps-test-message-#{SecureRandom.uuid}" }

    # using std ruby thread utils not to introduce any new deps:
    consumed = Queue.new # read_messages result queue (Array[MSG])
    consumer = Thread.new { consumed << read_messages(topic_name, messages.count) }
    sleep 0.5 # give consumer thread a bit to start + actually subscribe

    topic = subject.topic(topic_name)
    messages.each {|msg| topic.publish(msg) }
    subject.close

    consumer.join(10) # wait for consumer thread to finish
    expect(consumed.size).to eq(1) # make sure it returned smth so .pop will not block
    expect(consumed.pop).to match_array(messages)
  end
end
