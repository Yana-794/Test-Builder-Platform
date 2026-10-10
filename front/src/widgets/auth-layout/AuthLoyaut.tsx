import type { FC, ReactNode } from "react";
import styles from './AuthLoyaut.module.css'
import { Link } from "react-router";
import { ArrowLeft } from "lucide-react";
import { Logo } from '../../shared/index';

type AuthLoyautProps = {
    title: string;
    subtitle: string;
    children?: ReactNode;
    footer?: ReactNode;
}

export const AuthLoyaut: FC<AuthLoyautProps> = ({title, subtitle, children, footer}) => {
return (
    <div className={styles.page}>
    <div className={styles.toBar}>
        <Link to="/" className={styles.backLink}>
          <ArrowLeft size={16} strokeWidth={2} />
          <span>Вернуться на главную</span>
        </Link>
    </div>
     <div className={styles.content}>
        <div className={styles.logoWrap}>
            <Logo variant="default"/>
        </div>
        <div className={styles.card}>
            <div className={styles.head}>
                <h1 className={styles.title}>
                    {title}
                </h1>
                {subtitle && (
                    <p className={styles.subtitle}>{subtitle}</p>
                )}
            </div>

            <div className={styles.body}>
                {children}
            </div>
            {footer && (
                <div className={styles.footer}>{footer}</div>
            )}
        </div>
     </div>
    </div>
)
}