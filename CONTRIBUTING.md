# Contributing

First, thank you for your interest in this project! I would be happy to see you contribute.

## Build

For convenience, everything is managed with Docker. Follow these steps to locally build the website:

- **Create a `.env` file**: You can make one from the example by running `cp env.example .env`.
- **Fill the `.env` file**: Provide values for `SESSION_SECRET` and `POSTGRES_PASSWORD`.
  - If you don't want to set up Google login, use the **dev** branch which bypasses authentication. 
  - Otherwise, replace `GOOGLE_ID` and `GOOGLE_SECRET` with your Google authentication credentials, which can be found in the Google API Dashboard.
- **Run Docker**: Execute `docker-compose -f compose-local.yml up --no-deps --build`.
- **Access the Website**: Connect to `http://localhost:8080` to view the site.

## Pull Requests (PR)

If you want to resolve an issue:

- Please **comment on the issue** as soon as possible to avoid duplicate work. If you have questions, you can discuss them in the issue or create a draft PR.

**Guidelines for PRs:**
- Ensure the PR title is descriptive and relevant.
- Try to keep commits clean and meaningful if possible.
