import { Routes } from '@angular/router';

import { LoginComponent } from './views/login/login.component';
import { EnvironmentComponent } from './views/environment/environment.component';

export const routes: Routes = [
  { path: '', redirectTo: 'login', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'environments', component: EnvironmentComponent }
];
