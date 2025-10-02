// Подключаем библиотеки (как require в PHP)
const { Pool } = require("pg");
const { v4: uuidv4 } = require("uuid");
require("dotenv").config();

// Функция для создания базы данных
async function initDatabase() {
  console.log("Начинаем инициализацию базы данных...");

  // Сначала подключаемся к стандартной базе postgres
  const adminPool = new Pool({
    host: process.env.DB_HOST,
    port: process.env.DB_PORT,
    database: "postgres",
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
  });

  try {
    const adminClient = await adminPool.connect();
    console.log("Подключение к PostgreSQL установлено");

    // Создаем базу данных (если не существует)
    try {
      await adminClient.query("CREATE DATABASE employee_management");
      console.log("База данных создана");
    } catch (error) {
      if (error.code === "42P04") {
        console.log("База данных уже существует");
      } else {
        throw error;
      }
    } finally {
      adminClient.release();
    }

    // Закрываем соединение с админ-пулом
    await adminPool.end();

    // Теперь подключаемся к нашей новой базе
    const pool = new Pool({
      host: process.env.DB_HOST,
      port: process.env.DB_PORT,
      database: process.env.DB_NAME,
      user: process.env.DB_USER,
      password: process.env.DB_PASSWORD,
    });

    const client = await pool.connect();

    console.log("Создаем таблицы...");

    // Таблица департаментов
    await client.query(`
      CREATE TABLE IF NOT EXISTS departments (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL UNIQUE,
        description TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      );
    `);
    console.log("Таблица departments создана");

    // Таблица сотрудников
    await client.query(`
      CREATE TABLE IF NOT EXISTS employees (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        full_name VARCHAR(255) NOT NULL,
        gender VARCHAR(10) NOT NULL CHECK (gender IN ('male', 'female')),
        age INTEGER NOT NULL CHECK (age >= 18 AND age <= 70),
        education VARCHAR(20) NOT NULL CHECK (education IN ('secondary', 'specialized', 'higher')),
        position VARCHAR(255) NOT NULL,
        passport VARCHAR(20) NOT NULL UNIQUE CHECK (passport ~ '^\\d{4} \\d{6}$'),
        department_id UUID NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
        status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'vacation', 'fired')),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        fired_at TIMESTAMP,
        vacation_start_at TIMESTAMP,
        vacation_end_at TIMESTAMP
      );
    `);
    console.log("Таблица employees создана");

    // Создаем индексы для быстрого поиска
    await client.query(`
      CREATE INDEX IF NOT EXISTS idx_employees_department_id ON employees(department_id);
      CREATE INDEX IF NOT EXISTS idx_employees_status ON employees(status);
      CREATE INDEX IF NOT EXISTS idx_employees_position ON employees(position);
      CREATE INDEX IF NOT EXISTS idx_employees_full_name ON employees(full_name);
    `);
    console.log("Индексы созданы");

    // Добавляем демо-данные
    console.log("Добавляем демо-данные...");

    // Департаменты
    const departments = [
      {
        id: uuidv4(),
        name: "IT-департамент",
        description: "Отдел информационных технологий",
      },
      {
        id: uuidv4(),
        name: "Отдел продаж",
        description: "Отдел по работе с клиентами и продажам",
      },
      {
        id: uuidv4(),
        name: "HR-отдел",
        description: "Отдел кадров и управления персоналом",
      },
      {
        id: uuidv4(),
        name: "Финансовый отдел",
        description: "Отдел финансов и бухгалтерии",
      },
      {
        id: uuidv4(),
        name: "Маркетинг",
        description: "Отдел маркетинга и рекламы",
      },
    ];

    for (const dept of departments) {
      await client.query(
        `INSERT INTO departments (id, name, description) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`,
        [dept.id, dept.name, dept.description]
      );
    }

    // Получаем ID департаментов для сотрудников
    const deptResult = await client.query(
      "SELECT id, name FROM departments ORDER BY name"
    );
    const deptMap = {};
    deptResult.rows.forEach((row) => {
      deptMap[row.name] = row.id;
    });

    // Добавляем сотрудников
    const employees = [
      {
        full_name: "Иванов Иван Иванович",
        gender: "male",
        age: 35,
        education: "higher",
        position: "Программист",
        passport: "1234 567890",
        department_id: deptMap["IT-департамент"],
        status: "active",
      },
      {
        full_name: "Петрова Анна Сергеевна",
        gender: "female",
        age: 28,
        education: "higher",
        position: "Аналитик",
        passport: "2345 678901",
        department_id: deptMap["IT-департамент"],
        status: "vacation",
        vacation_start_at: new Date(),
      },
      {
        full_name: "Сидоров Петр Александрович",
        gender: "male",
        age: 42,
        education: "higher",
        position: "Менеджер по продажам",
        passport: "3456 789012",
        department_id: deptMap["Отдел продаж"],
        status: "active",
      },
      {
        full_name: "Козлова Мария Викторовна",
        gender: "female",
        age: 31,
        education: "higher",
        position: "HR-менеджер",
        passport: "4567 890123",
        department_id: deptMap["HR-отдел"],
        status: "active",
      },
      {
        full_name: "Николаев Дмитрий Олегович",
        gender: "male",
        age: 38,
        education: "specialized",
        position: "Бухгалтер",
        passport: "5678 901234",
        department_id: deptMap["Финансовый отдел"],
        status: "fired",
        fired_at: new Date(),
      },
    ];

    for (const emp of employees) {
      await client.query(
        `
        INSERT INTO employees (
          full_name, gender, age, education, position, passport, 
          department_id, status, vacation_start_at, fired_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (passport) DO NOTHING
      `,
        [
          emp.full_name,
          emp.gender,
          emp.age,
          emp.education,
          emp.position,
          emp.passport,
          emp.department_id,
          emp.status,
          emp.vacation_start_at || null,
          emp.fired_at || null,
        ]
      );
    }

    console.log("Демо-данные добавлены");
    console.log("База данных готова к использованию!");

    client.release();
    await pool.end();
  } catch (error) {
    console.error("Ошибка инициализации базы данных:", error.message);
  }
}

// Запускаем инициализацию
initDatabase();
