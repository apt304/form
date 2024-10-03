# Contributing to Form

Thank you for considering contributing to the `form` library! We welcome all types of contributions, from simple bug fixes to major improvements. This guide will help you get started.

## Getting Started

1. **Fork the Repository**: Start by forking the [repository](https://github.com/apt304/form) to your GitHub account.

2. **Clone the Repository**: Clone your forked repository to your local machine.

   ```sh
   git clone https://github.com/<your-username>/form.git
   cd form
   ```

3. **Install Dependencies**: Make sure you have Go installed (version 1.16 or later). Install any additional dependencies if needed.

4. **Create a Branch**: Create a new branch for your contribution.

   ```sh
   git checkout -b my-feature-branch
   ```

## Making Changes

- **Code Style**: Please follow the standard Go coding conventions. We use `gofmt` to ensure consistency in formatting.
- **Tests**: Ensure that you write tests for any new functionality or changes you make.
- **Commit Messages**: Write clear and concise commit messages. Use the following format where appropriate:
    - `feat:` for new features.
    - `fix:` for bug fixes.
    - `docs:` for documentation updates.
    - `test:` for testing-related commits.
    - `refactor:` for code improvements that do not add or remove functionality.

## Running Tests

To run the tests, use the following command:

```sh
go test ./...
```

Make sure all tests pass before submitting your pull request.

## Submitting Your Contribution

1. **Push to GitHub**: Push your branch to your forked repository.

   ```sh
   git push origin my-feature-branch
   ```

2. **Open a Pull Request**: Navigate to the original repository and open a pull request. Make sure to provide a detailed description of the changes you made.

3. **Address Feedback**: The project maintainers may provide feedback or request changes to your pull request. Please address these comments as needed.

We appreciate all pull requests and suggestions, but please understand that not all proposed changes will be merged into the library. The project maintainers will review each submission and decide based on the library's goals and overall direction.

## Reporting Issues

If you encounter any bugs or have feature requests, please open an issue on the [GitHub repository](https://github.com/apt304/form/issues). Provide as much detail as possible, including steps to reproduce the issue.

## Code of Conduct

We follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/) to create a welcoming and inclusive environment for everyone. Please read and adhere to it during your participation in this project.

## Thank You!

Your contributions make this project better! Thank you for taking the time to help improve `form`.
