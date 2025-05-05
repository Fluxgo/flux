package flux

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/glebarez/sqlite" 
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
)

type Database struct {
	DB *gorm.DB
}

type DatabaseConfig struct {
	Driver        string        
	Name          string        
	Host          string        
	Port          int           
	Username      string        
	Password      string        
	SSLMode       string        
	Charset       string        
	Timezone      string        
	MaxOpenConns  int           
	MaxIdleConns  int           
	ConnMaxLife   time.Duration 
	SlowThreshold time.Duration 
	LogLevel      logger.LogLevel 
	Debug         bool          
}


func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:        "sqlite",
		Name:          "flux.db",
		Host:          "localhost",
		Port:          3306, 
		Charset:       "utf8mb4",
		Timezone:      "Local",
		MaxOpenConns:  100,
		MaxIdleConns:  10,
		ConnMaxLife:   time.Hour,
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Debug:         false,
	}
}

func NewDatabase(config *DatabaseConfig) (*Database, error) {
	if config == nil {
		config = DefaultDatabaseConfig()
	}

	if config.SlowThreshold == 0 {
		config.SlowThreshold = 200 * time.Millisecond
	}
	
	if config.Charset == "" {
		config.Charset = "utf8mb4"
	}
	
	if config.Timezone == "" {
		config.Timezone = "Local"
	}

	if config.Driver == "sqlite" {
		dbDir := filepath.Dir(config.Name)
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// GORM logger
	logLevel := config.LogLevel
	if logLevel == 0 {
		logLevel = logger.Info
	}
	
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             config.SlowThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	
	var dialector gorm.Dialector
	var err error
	
	switch config.Driver {
	case "sqlite":
		dialector = sqlite.Open(config.Name)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=%s",
			config.Username, config.Password, config.Host, config.Port, config.Name,
			config.Charset, config.Timezone)
		dialector = mysql.Open(dsn)
	case "postgres":
		sslMode := config.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
			config.Host, config.Port, config.Username, config.Password, config.Name, 
			sslMode, config.Timezone)
		dialector = postgres.Open(dsn)
	case "sqlserver":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
			config.Username, config.Password, config.Host, config.Port, config.Name)
		dialector = sqlserver.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	
	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	}
	
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	}
	
	if config.ConnMaxLife > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLife)
	}

	
	if config.Debug {
		db = db.Debug()
	}

	return &Database{DB: db}, nil
}


func (d *Database) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}


func (d *Database) Create(value interface{}) error {
	return d.DB.Create(value).Error
}


func (d *Database) First(dest interface{}, cond ...interface{}) error {
	return d.DB.First(dest, cond...).Error
}


func (d *Database) Find(dest interface{}, cond ...interface{}) error {
	return d.DB.Find(dest, cond...).Error
}


func (d *Database) Update(value interface{}) error {
	return d.DB.Save(value).Error
}


func (d *Database) Delete(value interface{}) error {
	return d.DB.Delete(value).Error
}


func (d *Database) Where(query interface{}, args ...interface{}) *gorm.DB {
	return d.DB.Where(query, args...)
}


func (d *Database) Transaction(fc func(tx *gorm.DB) error) error {
	return d.DB.Transaction(fc)
}


func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}


func (d *Database) Exec(sql string, values ...interface{}) error {
	return d.DB.Exec(sql, values...).Error
}

func (d *Database) Raw(sql string, dest interface{}, values ...interface{}) error {
	return d.DB.Raw(sql, values...).Scan(dest).Error
}


func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}


func (d *Database) GetDriverName() string {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return "unknown"
	}
	
	driverName := ""
	sqlDB.QueryRow("SELECT current_database()").Scan(&driverName)
	if driverName != "" {
		return "postgres"
	}
	
	// Use MySQL
	sqlDB.QueryRow("SELECT DATABASE()").Scan(&driverName)
	if driverName != "" {
		return "mysql"
	}
	
	// Use SQLite
	var version string
	sqlDB.QueryRow("SELECT sqlite_version()").Scan(&version)
	if version != "" {
		return "sqlite"
	}
	
	// Use SQL Server
	sqlDB.QueryRow("SELECT DB_NAME()").Scan(&driverName)
	if driverName != "" {
		return "sqlserver"
	}
	
	return "unknown"
}

func (d *Database) Model(value interface{}) *gorm.DB {
	return d.DB.Model(value)
}

func (d *Database) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return d.DB.Scopes(funcs...)
}

func (d *Database) Preload(query string, args ...interface{}) *gorm.DB {
	return d.DB.Preload(query, args...)
}


type Migration struct {
	Name      string
	Up        func(*gorm.DB) error
	Down      func(*gorm.DB) error
}


type MigrationManager struct {
	DB         *Database
	Migrations []Migration
}


func NewMigrationManager(db *Database) *MigrationManager {
	return &MigrationManager{
		DB:         db,
		Migrations: []Migration{},
	}
}

func (m *MigrationManager) AddMigration(name string, up, down func(*gorm.DB) error) {
	m.Migrations = append(m.Migrations, Migration{
		Name: name,
		Up:   up,
		Down: down,
	})
}


func (m *MigrationManager) Migrate() error {
	
	err := m.DB.DB.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`).Error
	
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	var appliedMigrations []string
	err = m.DB.DB.Raw("SELECT name FROM migrations").Scan(&appliedMigrations).Error
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	
	appliedMap := make(map[string]bool)
	for _, name := range appliedMigrations {
		appliedMap[name] = true
	}
	
	for _, migration := range m.Migrations {
		if !appliedMap[migration.Name] {
			err := m.DB.Transaction(func(tx *gorm.DB) error {
				if err := migration.Up(tx); err != nil {
					return err
				}
				
				
				return tx.Exec("INSERT INTO migrations (name) VALUES (?)", migration.Name).Error
			})
			
			if err != nil {
				return fmt.Errorf("failed to apply migration '%s': %w", migration.Name, err)
			}
			
			log.Printf("Applied migration: %s", migration.Name)
		}
	}
	
	return nil
}


func (m *MigrationManager) Rollback(steps int) error {
	
	var appliedMigrations []string
	err := m.DB.DB.Raw("SELECT name FROM migrations ORDER BY id DESC LIMIT ?", steps).Scan(&appliedMigrations).Error
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	
	migrationMap := make(map[string]Migration)
	for _, migration := range m.Migrations {
		migrationMap[migration.Name] = migration
	}
	
	
	for _, name := range appliedMigrations {
		migration, ok := migrationMap[name]
		if !ok {
			return fmt.Errorf("migration '%s' not found", name)
		}
		
		
		err := m.DB.Transaction(func(tx *gorm.DB) error {
			
			if err := migration.Down(tx); err != nil {
				return err
			}
			
			
			return tx.Exec("DELETE FROM migrations WHERE name = ?", name).Error
		})
		
		if err != nil {
			return fmt.Errorf("failed to roll back migration '%s': %w", name, err)
		}
		
		log.Printf("Rolled back migration: %s", name)
	}
	
	return nil
}
