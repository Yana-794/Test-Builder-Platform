import { createBrowserRouter } from "react-router";
import { LandingPage } from "../../pages/landing/LandingPage";
import { LoginPage } from "../../pages/auth/login/LoginPage";
import { RegisterPage } from "../../pages/auth/register/RegisterPage";
import { VerifyEmailPage } from "../../pages/auth/verify-email/VerifyEmailPage";
import { NotFoundPage } from "../../pages/not-found/NotFoundPage";
import { DashboardPage } from "../../pages/dashboard/DashboardPage";
import { ProtectedRoute } from "./ProtectedRoute";
import { AppLayout } from "../../widgets/app-layout/AppLayout";
import { TestsPage } from "../../pages/test-list-page/TestsPage";
import { TestCreatePage } from "../../pages/test-create-page/TestCreatePage";
import { TestEditorPage } from "../../pages/test-editor-page/TestEditorPage";

export const router = createBrowserRouter([
  { path: "/", element: <LandingPage /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  { path: "/verify-email", element: <VerifyEmailPage /> },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          { path: "/app", element: <DashboardPage /> },
          { path: "/app/tests", element: <TestsPage /> },
          { path: "/app/tests/new", element: <TestCreatePage /> },
          { path: "/app/tests/:testId/edit", element: <TestEditorPage /> },
          // { path: '/app/students', element: <StudentsPage /> },
          // { path: '/app/results', element: <ResultsPage /> },
          // { path: '/app/profile', element: <ProfilePage /> },
        ],
      },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
]);
