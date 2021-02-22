import {Injectable} from '@angular/core';
import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Observable, throwError} from 'rxjs';
import {catchError} from 'rxjs/operators';

import {AuthService} from '@services/auth.service';
import {ToastrService} from 'ngx-toastr';

@Injectable()
export class ErrorInterceptor implements HttpInterceptor {
    constructor(private authenticationService: AuthService, private toaster: ToastrService) {
    }

    intercept(request: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        return next.handle(request).pipe(catchError(err => {
            let message;
            if (err.status === 401 && location.pathname !== '/login') {
                // auto logout if 401 response returned from api
                this.authenticationService.logout();
                setTimeout(() => location.pathname = '/login', 100);
            } else {
                message = message = !!err.error ? err.error.message : err.statusText;
                err.status !== 500 ?
                    this.toaster.warning(message, 'Attention') :
                    this.toaster.error(message, 'Achtung!');
            }

            return throwError(message);
        }));
    }
}
