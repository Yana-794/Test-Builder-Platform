import { Link } from "react-router";
import { AuthLoyaut } from "../../../widgets/auth-layout/AuthLoyaut";
import { RegisterForm } from "../../../features/auth/ui/register-form/ui/RegisterForm";

export const RegisterPage = () => {
  return (
    <AuthLoyaut
      title="Регистрация"
      subtitle="Создайте новый аккаунт"
      footer={
        <>
          Уже есть аккаунт ? <Link to="/login">Войти</Link>
        </>
      }
    >
      <RegisterForm />
    </AuthLoyaut>
  );
};
