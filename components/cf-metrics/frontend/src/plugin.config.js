(function () {
  'use strict';

  // register this plugin application to the platform
  if (env && env.registerApplication) {
    env.registerApplication(
      'cfMetrics',            // plugin application identity
      'cf-metrics',           // plugin application's root angular module name
      'cf-metrics/',         // plugin application's base path
      ''            // plugin applications's start state
    );
  }

}());
