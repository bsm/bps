require 'spec_helper'

RSpec.describe BPS::Publisher::InMem do
  it 'should maintain topics' do
    expect(subject.topic('x')).to be_a(described_class::Topic)
    expect(subject.topic_names).to eq(%w[x])
    expect(subject.topic('z')).to be_a(described_class::Topic)
    expect(subject.topic_names).to eq(%w[x z])
    expect(subject.topic('y')).to be_a(described_class::Topic)
    expect(subject.topic_names).to eq(%w[x y z])
  end

  it 'should publish' do
    topic = subject.topic('x')
    topic.publish('foo')
    topic.publish('bar')
    expect(topic.messages).to eq(%w[foo bar])

    expect { topic.flush }.not_to raise_error
    expect(subject.topic('x').messages).to eq(%w[foo bar])
  end
end
