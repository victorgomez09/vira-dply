import { inject } from "@angular/core";
import { Auth } from "../services/auth/auth";
import { HttpHandlerFn, HttpInterceptorFn, HttpRequest } from "@angular/common/http";

export const jwtInterceptor: HttpInterceptorFn = (req: HttpRequest<any>, next: HttpHandlerFn) => {
  const authService = inject(Auth);
  const token = authService.getToken();

  if (token) {
    const cloned = req.clone({
      setHeaders: { Authorization: `Bearer ${token}` }
    });
    return next(cloned);
  }

  return next(req);
};