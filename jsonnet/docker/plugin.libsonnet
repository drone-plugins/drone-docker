local windows_pipe = '\\\\\\\\.\\\\pipe\\\\docker_engine';
local windows_pipe_volume = 'docker_pipe';

local windows_settings = {
  daemon_off: true,
  // Workaround for https://github.com/drone/drone-cli/issues/117
  purge: 'false',
};

{
  /**
   * Returns a Docker plugin using the given settings.
   *
   * @param name The name of the pipeline step.
   * @param settings The settings for the plugin.
   * @return A json representing a Docker plugin and its settings.
   */
  Plugin(name, settings): {
    name: name,
    image: 'plugins/docker',

    settings: settings,
  },

  /**
   * Returns a Docker plugin for Windows builds.
   *
   * Currently Windows does not support docker in docker so the pipe must be
   * mounted as a volume. This requires trusted builds to be enabled for the
   * repository.
   *
   * The pipeline must have
   *
   * @param name The name of the pipeline step.
   * @param settings The settings for the plugin.
   * @return A json representing a Docker plugin and its settings.
   */
  WindowsPlugin(name, settings): self.Plugin(name, settings + windows_settings) {
    volumes: [{ name: windows_pipe_volume, path: windows_pipe }],
  },

  /**
   * The host volume for Window's docker pipe.
   *
   * This needs to be applied to the pipeline when using the Docker plugin on
   * Windows.
   */
  WindowsHostVolume: {
    name: windows_pipe_volume,
    host: { path: windows_pipe },
  },

  /**
   * Returns the default settings for the Docker plugin.
   *
   * @param repo The repository of the Docker image.
   * @param dockerfile The dockerfile to use.
   * @param context The directory to run the build from.
   * @return A json representing the Docker plugin's settings.
   */
  Settings(repo, dockerfile='Dockerfile', context=''): {
    repo: repo,
    dockerfile: dockerfile,
    context: context,
  },

  /**
   * Returns the settings for automatically tagging the Docker image.
   *
   * @param suffix The suffix to append to the image tag.
   * @return The settings for automatically tagging a Docker image.
   */
  AutoTag(suffix): {
    auto_tag: true,
    auto_tag_suffix: suffix,
  },

  /**
   * Returns the settings for adding build arguments.
   *
   * Build arguments should be done in the form `FOO=foo`.
   *
   * @param args A list of build arguments.
   * @return The settings for build arguments for a Docker image.
   */
  BuildArguments(args): {
    build_args: args,
  },

  /**
   * Disables pulling of images during a build.
   *
   * This should only be used when mounting the docker pipe.
   */
  DisablePull: {
    // Workaround for https://github.com/drone/drone-cli/issues/117
    pull_image: 'false',
  },

  /**
   * Squashes the docker image.
   */
  Squash: {
    squash: true,
  },

  /**
   * Returns authentication settings for pusing the Docker image.
   *
   * @param username The name of the secret containing the username for the Docker registry.
   * @param password The name of the secret containing the password for the Docker registry.
   */
  Authenticate(username='docker_username', password='docker_password'): {
    username: { from_secret: username },
    password: { from_secret: password },
  },
}
