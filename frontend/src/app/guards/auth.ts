import { inject } from "@angular/core";
import { CanActivateFn, Router } from "@angular/router";
import { Auth } from "../services/auth/auth";

export const authGuard: CanActivateFn = (route, state) => {
  const authService = inject(Auth);
  const router = inject(Router);

  if (authService.isLoggedIn()) {
    return true; // Usuario logeado, permite acceder
  } else {
    router.navigate(['/login']); // Redirige al login
    return false;
  }
};