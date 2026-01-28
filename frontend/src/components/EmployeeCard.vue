<template>
  <div class="employee-card" :class="{ fired: employee.status === 'fired' }">
    <div class="employee-header">
      <div class="employee-name">{{ employee.full_name }}</div>
      <div class="employee-status" :class="`status-${employee.status}`">
        {{ getStatusText(employee.status) }}
      </div>
    </div>
    <div class="employee-info">
      <div class="info-item">
        <div class="info-label">Должность</div>
        <div class="info-value">{{ employee.position }}</div>
      </div>
      <div class="info-item">
        <div class="info-label">Пол / Возраст</div>
        <div class="info-value">
          {{ employee.gender === "male" ? "М" : "Ж" }} / {{ employee.age }} лет
        </div>
      </div>
      <div class="info-item">
        <div class="info-label">Образование</div>
        <div class="info-value">{{ getEducationText(employee.education) }}</div>
      </div>
      <div class="info-item">
        <div class="info-label">Паспорт</div>
        <div class="info-value">{{ employee.passport }}</div>
      </div>
    </div>
    <div class="employee-actions">
      <button
        class="btn btn-primary"
        @click="$emit('edit', employee)"
        v-if="employee.status !== 'fired'"
        :disabled="loading"
      >
        Редактировать
      </button>
      <button
        class="btn btn-success"
        @click="$emit('change-status', 'vacation')"
        v-if="employee.status === 'active'"
        :disabled="loading"
      >
        В отпуск
      </button>
      <button
        class="btn btn-success"
        @click="$emit('change-status', 'active')"
        v-if="employee.status === 'vacation'"
        :disabled="loading"
      >
        Вернуть
      </button>
      <button
        class="btn btn-danger"
        @click="$emit('fire', employee)"
        v-if="employee.status !== 'fired'"
        :disabled="loading"
      >
        Уволить
      </button>
    </div>
  </div>
</template>

<script setup>
import { defineProps, defineEmits } from "vue";

defineProps({
  employee: Object,
  loading: Boolean,
});

defineEmits(["edit", "change-status", "fire"]);

const getStatusText = (status) => {
  const statusMap = {
    active: "Активен",
    vacation: "В отпуске",
    fired: "Уволен",
  };
  return statusMap[status] || status;
};

const getEducationText = (education) => {
  const educationMap = {
    secondary: "Среднее",
    specialized: "Средне-специальное",
    higher: "Высшее",
  };
  return educationMap[education] || education;
};
</script>

<style scoped>
.employee-card {
  background: white;
  border-radius: 10px;
  padding: 25px;
  box-shadow: 0 5px 15px rgba(0, 0, 0, 0.08);
  border: 1px solid #f0f0f0;
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}

.employee-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
}

.employee-card.fired {
  opacity: 0.6;
  background: #ffffff;
  border-color: #dc3545;
}

.employee-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 15px;
  flex-wrap: wrap;
  gap: 10px;
}

.employee-name {
  font-size: 1.3em;
  font-weight: bold;
  color: #2c3e50;
  flex: 1;
}

.employee-status {
  padding: 5px 12px;
  border-radius: 15px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.status-active {
  background: #d4edda;
  color: #155724;
}

.status-vacation {
  background: #fff3cd;
  color: #856404;
}

.status-fired {
  background: #f8d7da;
  color: #721c24;
}

.employee-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
}

.info-label {
  font-size: 12px;
  color: #6c757d;
  text-transform: uppercase;
  font-weight: 600;
  margin-bottom: 5px;
}

.info-value {
  color: #2c3e50;
  font-weight: 500;
}

.employee-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.btn {
  padding: 12px 25px;
  border: none;
  border-radius: 5px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: all 0.3s ease;
  text-decoration: none;
  display: inline-block;
  text-align: center;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background-color: #0086a1;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(74, 74, 74, 0.4);
}

.btn-danger {
  background-color: #bb4646;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(255, 107, 107, 0.4);
}

.btn-success {
  background-color: #38be6f;
  color: white;
}

.btn-success:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px #2ecc7166;
}
</style>
