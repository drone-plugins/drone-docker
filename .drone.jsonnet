local pipeline = import 'pipeline.libsonnet';

[
  pipeline.test('linux', 'amd64'),
  pipeline.build('docker', 'linux', 'amd64')
  pipeline.build('gcr', 'linux', 'amd64')
]
