import {Injectable} from '@angular/core';
import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Observable} from 'rxjs';

import {environment} from '@env/environment';
import {AuthService} from '@services/auth.service';

@Injectable()
export class JwtInterceptor implements HttpInterceptor {
    constructor(private authenticationService: AuthService) {
    }

    intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        // add auth header with jwt if user is logged in and request is to the api url
        const session = this.authenticationService.currentSessionValue;
        if ((session && session.token) && request.url.startsWith(environment.apiUrl)) {
            request = request.clone({
                setHeaders: {Authorization: `Bearer ${session.token}`}
            });
        }

        return next.handle(request);
    }
}
