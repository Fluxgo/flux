# testproject

A web application built with flux Framework.

## Getting Started

1. Run the development server:
   
   ```bash
   flux serve
   ```

2. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Database Configuration

This project uses SQLite by default, which requires no additional setup. To use other databases:

1. Edit the database configuration in `config/flux.yaml`
2. Choose from: sqlite, mysql, postgres, sqlserver
3. Provide connection details as required

## Creating Controllers and Models

Generate new controllers:

```bash
flux generate controller User
```

Generate new models:

```bash
flux generate model User
```

## Learn More

To learn more about flux Framework, check out the documentation at flux Framework Documentation(https://github.com/Fluxgo/flux).
