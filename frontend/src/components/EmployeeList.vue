<template>
  <div class="employee-list-container">
    <div v-if="loading" class="loading">Загрузка сотрудников...</div>
    <div v-else-if="!selectedDepartmentId" class="empty-state">
      <h3>Выберите департамент</h3>
      <p>Для просмотра списка сотрудников необходимо выбрать департамент</p>
    </div>
    <div v-else-if="employees.length === 0" class="empty-state">
      <h3>Сотрудники не найдены</h3>
      <p>В выбранном департаменте пока нет сотрудников</p>
    </div>
    <div v-else class="employee-list">
      <EmployeeCard
        v-for="employee in employees"
        :key="employee.id"
        :employee="employee"
        @edit="$emit('edit', employee)"
        @change-status="$emit('change-status', employee.id, $event)"
        @fire="$emit('fire', employee)"
      />
    </div>
  </div>
</template>

<script setup>
import EmployeeCard from "./EmployeeCard.vue";

defineProps({
  selectedDepartmentId: String,
  loading: Boolean,
  employees: Array,
});

defineEmits(["edit", "change-status", "fire"]);
</script>

<style scoped>
.employee-list-container {
  width: 100%;
}

.loading {
  text-align: center;
  padding: 20px;
  color: #667eea;
  font-weight: 600;
}

.empty-state {
  text-align: center;
  padding: 50px 20px;
  color: #6c757d;
}

.empty-state h3 {
  margin-bottom: 15px;
  color: #2c3e50;
}

.employee-list {
  display: grid;
  gap: 20px;
}

/* Стили для мобильных устройств */
@media (max-width: 768px) {
  .employee-list {
    gap: 15px;
  }

  .empty-state {
    padding: 30px 15px;
  }
}
</style>
