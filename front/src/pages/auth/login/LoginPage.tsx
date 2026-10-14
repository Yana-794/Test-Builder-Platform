import { Link } from "react-router";
import { AuthLoyaut } from "../../../widgets/auth-layout/AuthLoyaut";
import { LoginForm } from "../../../features/auth/ui/login-form/ui/LoginForm";

export const LoginPage = () => {
  return (
    <AuthLoyaut
      title="Вход в аккаунт"
      subtitle="Рады видеть Вас снова"
      footer={
        <>
          Нет аккаунта ? <Link to="/register">Зарегистрироваться</Link>
        </>
      }
    >
      <LoginForm />
    </AuthLoyaut>
  );
};
