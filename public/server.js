const express = require("express");
const { Pool } = require("pg");
const cors = require("cors");
const helmet = require("helmet");
const path = require("path");
require("dotenv").config();

const app = express();
const PORT = process.env.PORT || 3000;

// Подключаемся к базе данных
const pool = new Pool({
  host: process.env.DB_HOST,
  port: process.env.DB_PORT,
  database: process.env.DB_NAME,
  user: process.env.DB_USER,
  password: process.env.DB_PASSWORD,
});

// Middleware
app.use(helmet({ contentSecurityPolicy: false }));
app.use(cors());
app.use(express.json());
app.use(express.static(__dirname));

// ==================== API МАРШРУТЫ ====================

// GET /api/departments - Получить все департаменты
app.get("/api/departments", async (req, res) => {
  try {
    const result = await pool.query(
      "SELECT id, name, description FROM departments ORDER BY name"
    );
    res.json({
      success: true,
      data: result.rows,
    });
  } catch (error) {
    console.error("Ошибка получения департаментов:", error);
    res.status(500).json({
      success: false,
      error: "Ошибка получения департаментов",
    });
  }
});

// GET /api/employees/department/:departmentId - Сотрудники департамента
app.get("/api/employees/department/:departmentId", async (req, res) => {
  try {
    const { departmentId } = req.params;
    const result = await pool.query(
      `
      SELECT e.*, d.name as department_name
      FROM employees e
      JOIN departments d ON e.department_id = d.id
      WHERE e.department_id = $1
      ORDER BY e.full_name
    `,
      [departmentId]
    );

    res.json({
      success: true,
      data: result.rows,
    });
  } catch (error) {
    console.error("Ошибка получения сотрудников:", error);
    res.status(500).json({
      success: false,
      error: "Ошибка получения сотрудников",
    });
  }
});

// POST /api/employees/search - Поиск сотрудников
app.post("/api/employees/search", async (req, res) => {
  try {
    const { fullName, position, gender, education, ageFrom, ageTo } = req.body;

    let query = `
      SELECT e.*, d.name as department_name
      FROM employees e
      JOIN departments d ON e.department_id = d.id
      WHERE 1=1
    `;
    const params = [];
    let paramCount = 0;

    if (fullName) {
      paramCount++;
      query += ` AND e.full_name ILIKE $${paramCount}`;
      params.push(`%${fullName}%`);
    }

    if (position) {
      paramCount++;
      query += ` AND e.position = $${paramCount}`;
      params.push(position);
    }

    if (gender) {
      paramCount++;
      query += ` AND e.gender = $${paramCount}`;
      params.push(gender);
    }

    if (education) {
      paramCount++;
      query += ` AND e.education = $${paramCount}`;
      params.push(education);
    }

    if (ageFrom) {
      paramCount++;
      query += ` AND e.age >= $${paramCount}`;
      params.push(ageFrom);
    }

    if (ageTo) {
      paramCount++;
      query += ` AND e.age <= $${paramCount}`;
      params.push(ageTo);
    }

    query += " ORDER BY e.full_name";

    const result = await pool.query(query, params);
    res.json({
      success: true,
      data: result.rows,
    });
  } catch (error) {
    console.error("Ошибка поиска сотрудников:", error);
    res.status(500).json({
      success: false,
      error: "Ошибка поиска сотрудников",
    });
  }
});

// POST /api/employees - Создать сотрудника
app.post("/api/employees", async (req, res) => {
  try {
    const {
      fullName,
      gender,
      age,
      education,
      position,
      passport,
      departmentId,
    } = req.body;

    if (
      !fullName ||
      !gender ||
      !age ||
      !education ||
      !position ||
      !passport ||
      !departmentId
    ) {
      return res.status(400).json({
        success: false,
        error: "Все поля обязательны для заполнения",
      });
    }

    const result = await pool.query(
      `
      INSERT INTO employees (
        full_name, gender, age, education, position, passport, department_id
      ) VALUES ($1, $2, $3, $4, $5, $6, $7)
      RETURNING *
    `,
      [fullName, gender, age, education, position, passport, departmentId]
    );

    res.status(201).json({
      success: true,
      data: result.rows[0],
      message: "Сотрудник успешно создан",
    });
  } catch (error) {
    console.error("Ошибка создания сотрудника:", error);

    if (error.code === "23505") {
      res.status(400).json({
        success: false,
        error: "Сотрудник с таким паспортом уже существует",
      });
    } else {
      res.status(500).json({
        success: false,
        error: "Ошибка создания сотрудника",
      });
    }
  }
});

// PUT /api/employees/:id - Обновить сотрудника
app.put("/api/employees/:id", async (req, res) => {
  try {
    const { id } = req.params;
    const { fullName, gender, age, education, position, passport } = req.body;

    const result = await pool.query(
      `
      UPDATE employees 
      SET full_name = $1, gender = $2, age = $3, education = $4, 
          position = $5, passport = $6, updated_at = CURRENT_TIMESTAMP
      WHERE id = $7
      RETURNING *
    `,
      [fullName, gender, age, education, position, passport, id]
    );

    if (result.rows.length === 0) {
      return res.status(404).json({
        success: false,
        error: "Сотрудник не найден",
      });
    }

    res.json({
      success: true,
      data: result.rows[0],
      message: "Данные сотрудника обновлены",
    });
  } catch (error) {
    console.error("Ошибка обновления сотрудника:", error);
    res.status(500).json({
      success: false,
      error: "Ошибка обновления сотрудника",
    });
  }
});

// PATCH /api/employees/:id/status - Изменить статус сотрудника
app.patch("/api/employees/:id/status", async (req, res) => {
  try {
    const { id } = req.params;
    const { status } = req.body;

    let updateFields = "status = $1, updated_at = CURRENT_TIMESTAMP";
    let params = [status, id];

    if (status === "fired") {
      updateFields +=
        ", fired_at = CURRENT_TIMESTAMP, vacation_start_at = NULL, vacation_end_at = NULL";
    } else if (status === "vacation") {
      updateFields += ", vacation_start_at = CURRENT_TIMESTAMP";
    } else if (status === "active") {
      updateFields +=
        ", vacation_start_at = NULL, vacation_end_at = NULL, fired_at = NULL";
    }

    const result = await pool.query(
      `
      UPDATE employees 
      SET ${updateFields}
      WHERE id = $2
      RETURNING *
    `,
      params
    );

    if (result.rows.length === 0) {
      return res.status(404).json({
        success: false,
        error: "Сотрудник не найден",
      });
    }

    const statusMessages = {
      active: "Сотрудник активирован",
      vacation: "Сотрудник отправлен в отпуск",
      fired: "Сотрудник уволен",
    };

    res.json({
      success: true,
      data: result.rows[0],
      message: statusMessages[status] || "Статус изменен",
    });
  } catch (error) {
    console.error("Ошибка изменения статуса:", error);
    res.status(500).json({
      success: false,
      error: "Ошибка изменения статуса",
    });
  }
});

// GET /api/positions - Получить список должностей
app.get("/api/positions", async (req, res) => {
  try {
    const result = await pool.query(
      "SELECT DISTINCT position FROM employees ORDER BY position"
    );

    const basePositions = [
      "Программист",
      "Аналитик",
      "Тестировщик",
      "Менеджер по продажам",
      "HR-менеджер",
      "Бухгалтер",
      "Маркетолог",
      "Дизайнер",
      "Системный администратор",
      "Руководитель отдела",
      "Директор",
      "Специалист",
    ];

    const existingPositions = result.rows.map((row) => row.position);
    const allPositions = [
      ...new Set([...basePositions, ...existingPositions]),
    ].sort();

    res.json({
      success: true,
      data: allPositions,
    });
  } catch (error) {
    console.error("Ошибка получения должностей:", error);
    res.status(500).json({
      success: false,
      error: "Ошибка получения должностей",
    });
  }
});

// Тестовый маршрут
app.get("/api/test", (req, res) => {
  res.json({
    success: true,
    message: "Сервер работает!",
    timestamp: new Date().toISOString(),
  });
});

// Главная страница - ИСПРАВЛЕНО
app.get("/", (req, res) => {
  res.sendFile(path.join(__dirname, "index.html"));
});

// Обработка 404 для API
app.use("/api/*", (req, res) => {
  res.status(404).json({
    success: false,
    error: "API endpoint не найден",
  });
});

// Глобальный обработчик ошибок
app.use((err, req, res, next) => {
  console.error("Глобальная ошибка:", err);
  res.status(500).json({
    success: false,
    error: "Внутренняя ошибка сервера",
  });
});

// Запуск сервера
app.listen(PORT, () => {
  console.log(`Сервер запущен на http://localhost:${PORT}`);
  console.log(`API доступно по http://localhost:${PORT}/api/test`);
});

// Graceful shutdown
process.on("SIGINT", async () => {
  console.log("\n Завершение работы сервера...");
  await pool.end();
  process.exit(0);
});
