// eslint.config.js
module.exports = [
  {
    files: ["**/*.js", "**/*.jsx"],
    ignores: ["node_modules/**"],
    languageOptions: {
      ecmaVersion: 2021,
      sourceType: "module"
    },
    linterOptions: {
      reportUnusedDisableDirectives: true
    },
    rules: {
      "no-unused-vars": "warn",
      "no-console": "off"
    }
  }
];
