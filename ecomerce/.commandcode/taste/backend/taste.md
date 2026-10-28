# Backend / Infrastructure

- Prefers **Better Auth** as the authentication library (over hand-rolled auth or other providers). Confidence: 0.7
- Uses **PostgreSQL** as the database for real (non-mock) features such as auth sessions and user data. Confidence: 0.6
- Wants **placeholder environment variables**: commit a `.env.example` with placeholder values and keep the real `.env` gitignored — never commit real secrets. Confidence: 0.7
