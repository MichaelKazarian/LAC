;;; .dir-locals.el --- 

;; Copyright (C) Michael Kazarian
;;
;; Author: Michael Kazarian <michael.kazarian@gmail.com>
;; Keywords: 
;; Requirements: 
;; Status: not intended to be distributed yet

((nil . ((company-clang-arguments . (
                                     "-I/home/kazarian/.platformio/packages/framework-lgt8fx/libraries/Wire"))
         (eval . (progn
                   (add-to-list 'company-backends '(company-etags company-clang))))
         (eval . (message ".dir-locals.el was loaded"))
         )))

;;; .dir-locals.el ends here
