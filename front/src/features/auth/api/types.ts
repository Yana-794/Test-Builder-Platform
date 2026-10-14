export type User = {
  id: string;
  email: string;
  display_name: string;
  email_verified: boolean;
  created_at: string;
};
export type AuthResponse = {
  user: User;
  csrf_token: string;
};

export type RegisterPayload = {
  email: string;
  display_name: string;
  password: string;
};
export type LoginPayload = {
  email: string;
  password: string;
};

export type ApiError = { //TODO: зарегистрировать ошибку для всего приложения
  error: {
    code: string;
    message: string;
  };
};