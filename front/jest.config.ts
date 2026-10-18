import type { Config } from "jest";

const config: Config = {
  testEnvironment: "jsdom",
  setupFilesAfterEnv: ["<rootDir>/src/test/setup.ts"],
  transform: {
    "^.+\\.(ts|tsx)$": [
      "@swc/jest",
      {
        jsc: {
          parser: { syntax: "typescript", tsx: true },
          transform: { react: { runtime: "automatic" } },
        },
      },
    ],
  },
  moduleNameMapper: {
    "\\.(css|less|scss|module\\.css)$": "identity-obj-proxy",
    // если используешь vite env vars — замапь на заглушку
    "^import\\.meta\\.env$": "<rootDir>/src/test/envMock.ts",
  },
  testMatch: ["**/?(*.)+(test).(ts|tsx)"],
};

export default config;