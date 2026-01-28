export function useEmployeeApi() {
  const baseUrl = "/api";

  async function request(url, options = {}) {
    try {
      const response = await fetch(`${baseUrl}${url}`, {
        headers: {
          "Content-Type": "application/json",
          ...options.headers,
        },
        ...options,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || "Ошибка сервера");
      }

      return data;
    } catch (err) {
      throw new Error(err.message || "Ошибка сети");
    }
  }

  return {
    async get(url) {
      return request(url);
    },

    async post(url, data) {
      return request(url, {
        method: "POST",
        body: JSON.stringify(data),
      });
    },

    async put(url, data) {
      return request(url, {
        method: "PUT",
        body: JSON.stringify(data),
      });
    },

    async patch(url, data) {
      return request(url, {
        method: "PATCH",
        body: JSON.stringify(data),
      });
    },
  };
}
