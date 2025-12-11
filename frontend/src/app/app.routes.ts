import { Routes } from '@angular/router';
import { authGuard } from './guards/auth';

export const routes: Routes = [
    {
        path: 'login',
        loadComponent: () => import('./views/login/login').then(m => m.Login)
    },
    {
        path: 'dashboard',
        children: [
            {
                path: 'environments',
                loadComponent: () => import('./views/dashboard/environments/environments').then(m => m.Environments)
            }
        ],
        canActivate: [authGuard]
    }
];
