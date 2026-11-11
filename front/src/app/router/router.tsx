import { createBrowserRouter } from "react-router";
import { LandingPage } from "../../pages/landing/LandingPage";
import { LoginPage } from "../../pages/auth/login/LoginPage";
import { RegisterPage } from "../../pages/auth/register/RegisterPage";
import { NotFoundPage } from "../../pages/not-found/NotFoundPage";
import { DashboardPage } from "../../pages/dashboard/DashboardPage";
import { ProtectedRoute } from "./ProtectedRoute";
import { AppLayout } from "../../widgets/app-layout/AppLayout";
import { TestsPage } from "../../pages/test-list-page/TestsPage";
import { TestCreatePage } from "../../pages/test-create-page/TestCreatePage";
import { TestEditorPage } from "../../pages/test-editor-page/TestEditorPage";
import { TestDetailPage } from "../../pages/test-detail-page/TestDetailPage";
import { PublicLayout } from "../../widgets/public-layout/PublicLayout";
import { PublicTestStartPage } from "../../pages/public-test-start/PublicTestStartPage";
import { PublicAttemptPage } from "../../pages/public-attempt/PublicAttemptPage";
import { PublicResultPage } from "../../pages/public-result/PublicResultPage";
import { AnalyticsPage } from "../../pages/analytics-page/AnalyticsPage";

export const router = createBrowserRouter([
  { path: "/", element: <LandingPage /> },
  { path: "/login", element: <LoginPage /> },
  { path: "/register", element: <RegisterPage /> },
  
   {
    element: <PublicLayout />,
    children: [
      { path: '/test/:token', element: <PublicTestStartPage /> },
      { path: '/test/:token/attempt/:attemptId', element: <PublicAttemptPage /> },
      { path: '/test/:token/result/:attemptId', element: <PublicResultPage /> },
    ],
  },
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
          { path: "/app/tests/:testId", element: <TestDetailPage /> },
          // { path: '/app/students', element: <StudentsPage /> },
          { path: '/app/results', element: <AnalyticsPage /> },
          // { path: '/app/profile', element: <ProfilePage /> },
        ],
      },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
]);
